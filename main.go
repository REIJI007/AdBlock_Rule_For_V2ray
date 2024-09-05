package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

var (
	outputName = flag.String("outputname", "adblock.dat", "Name of the generated dat file")
	outputDir  = flag.String("outputdir", "./", "Directory to place the generated file")
)

type Entry struct {
	Type  string
	Value string
	Attrs []*router.Domain_Attribute
}

type List struct {
	Name  string
	Entry []Entry
}

type ParsedList struct {
	Name      string
	Inclusion map[string]bool
	Entry     []Entry
}

// 新增的函数，用于将 ParsedList 转换为纯文本（可选，如果需要）
func (l *ParsedList) toPlainText(listName string) error {
	var entryBytes []byte
	for _, entry := range l.Entry {
		var attrString string
		if entry.Attrs != nil {
			for _, attr := range entry.Attrs {
				attrString += "@" + attr.GetKey() + ","
			}
			attrString = strings.TrimRight(":"+attrString, ",")
		}
		entryBytes = append(entryBytes, []byte(entry.Type+":"+entry.Value+attrString+"\n")...)
	}
	if err := os.WriteFile(filepath.Join(*outputDir, listName+".txt"), entryBytes, 0644); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

func (l *ParsedList) toProto() (*router.GeoSite, error) {
	site := &router.GeoSite{
		CountryCode: l.Name,
	}
	for _, entry := range l.Entry {
		switch entry.Type {
		case "domain":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_RootDomain,
				Value:     entry.Value,
				Attribute: entry.Attrs,
			})
		case "regexp":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Regex,
				Value:     entry.Value,
				Attribute: entry.Attrs,
			})
		case "keyword":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Plain,
				Value:     entry.Value,
				Attribute: entry.Attrs,
			})
		case "full":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Full,
				Value:     entry.Value,
				Attribute: entry.Attrs,
			})
		default:
			return nil, errors.New("unknown domain type: " + entry.Type)
		}
	}
	return site, nil
}

func removeComment(line string) string {
	idx := strings.Index(line, "#")
	if idx == -1 {
		return line
	}
	return strings.TrimSpace(line[:idx])
}

func parseDomain(domain string, entry *Entry) error {
	kv := strings.Split(domain, ":")
	if len(kv) == 1 {
		entry.Type = "domain"
		entry.Value = strings.ToLower(kv[0])
		return nil
	}

	if len(kv) == 2 {
		entry.Type = strings.ToLower(kv[0])
		entry.Value = strings.ToLower(kv[1])
		return nil
	}

	return errors.New("Invalid format: " + domain)
}

func parseAttribute(attr string) (*router.Domain_Attribute, error) {
	var attribute router.Domain_Attribute
	if len(attr) == 0 || attr[0] != '@' {
		return &attribute, errors.New("invalid attribute: " + attr)
	}

	// Trim attribute prefix `@` character
	attr = attr[1:]
	parts := strings.Split(attr, "=")
	if len(parts) == 1 {
		attribute.Key = strings.ToLower(parts[0])
		attribute.TypedValue = &router.Domain_Attribute_BoolValue{BoolValue: true}
	} else {
		attribute.Key = strings.ToLower(parts[0])
		intv, err := strconv.Atoi(parts[1])
		if err != nil {
			return &attribute, errors.New("invalid attribute: " + attr + ": " + err.Error())
		}
		attribute.TypedValue = &router.Domain_Attribute_IntValue{IntValue: int64(intv)}
	}
	return &attribute, nil
}

func parseEntry(line string) (Entry, error) {
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")

	var entry Entry
	if len(parts) == 0 {
		return entry, errors.New("empty entry")
	}

	if err := parseDomain(parts[0], &entry); err != nil {
		return entry, err
	}

	for i := 1; i < len(parts); i++ {
		attr, err := parseAttribute(parts[i])
		if err != nil {
			return entry, err
		}
		entry.Attrs = append(entry.Attrs, attr)
	}

	return entry, nil
}

func Load(path string) (*List, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list := &List{
		Name: "adblock",
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = removeComment(line)
		if len(line) == 0 {
			continue
		}
		entry, err := parseEntry(line)
		if err != nil {
			return nil, err
		}
		list.Entry = append(list.Entry, entry)
	}

	return list, nil
}

func isMatchAttr(attrs []*router.Domain_Attribute, includeKey string) bool {
	isMatch := false
	mustMatch := true
	matchName := includeKey
	if strings.HasPrefix(includeKey, "!") {
		isMatch = true
		mustMatch = false
		matchName = strings.TrimLeft(includeKey, "!")
	}

	for _, attr := range attrs {
		attrName := attr.Key
		if mustMatch {
			if matchName == attrName {
				isMatch = true
				break
			}
		} else {
			if matchName == attrName {
				isMatch = false
				break
			}
		}
	}
	return isMatch
}

func createIncludeAttrEntrys(list *List, matchAttr *router.Domain_Attribute) []Entry {
	newEntryList := make([]Entry, 0, len(list.Entry))
	matchName := matchAttr.Key
	for _, entry := range list.Entry {
		matched := isMatchAttr(entry.Attrs, matchName)
		if matched {
			newEntryList = append(newEntryList, entry)
		}
	}
	return newEntryList
}

func ParseList(list *List, ref map[string]*List) (*ParsedList, error) {
	pl := &ParsedList{
		Name:      list.Name,
		Inclusion: make(map[string]bool),
	}
	entryList := list.Entry
	for {
		newEntryList := make([]Entry, 0, len(entryList))
		hasInclude := false
		for _, entry := range entryList {
			if entry.Type == "include" {
				refName := strings.ToUpper(entry.Value)
				if entry.Attrs != nil {
					for _, attr := range entry.Attrs {
						InclusionName := strings.ToUpper(refName + "@" + attr.Key)
						if pl.Inclusion[InclusionName] {
							continue
						}
						pl.Inclusion[InclusionName] = true

						refList := ref[refName]
						if refList == nil {
							return nil, errors.New(entry.Value + " not found.")
						}
						attrEntrys := createIncludeAttrEntrys(refList, attr)
						if len(attrEntrys) != 0 {
							newEntryList = append(newEntryList, attrEntrys...)
						}
					}
				} else {
					InclusionName := refName
					if pl.Inclusion[InclusionName] {
						continue
					}
					pl.Inclusion[InclusionName] = true
					refList := ref[refName]
					if refList == nil {
						return nil, errors.New(entry.Value + " not found.")
					}
					newEntryList = append(newEntryList, refList.Entry...)
				}
				hasInclude = true
			} else {
				newEntryList = append(newEntryList, entry)
			}
		}
		entryList = newEntryList
		if !hasInclude {
			break
		}
	}
	pl.Entry = entryList

	return pl, nil
}

func main() {
	flag.Parse()

	// 文件路径设置
	filePath := "./adblock.txt"
	outputFile := filepath.Join(*outputDir, *outputName)

	// 加载所有引用的列表，用于处理 include
	ref := make(map[string]*List)
	ref["ADBLOCK"] = nil // 预先定义，以防止自引用导致的问题

	// 读取 adblock.txt 文件内容
	fmt.Println("Reading domain list from", filePath)
	list, err := Load(filePath)
	if err != nil {
		fmt.Println("Failed to load file:", err)
		os.Exit(1)
	}
	ref["ADBLOCK"] = list

	// 解析列表，处理 include
	pl, err := ParseList(list, ref)
	if err != nil {
		fmt.Println("Failed to parse list:", err)
		os.Exit(1)
	}

	// 转换为 proto 格式并设置标签为 "adblock"
	site, err := pl.toProto()
	if err != nil {
		fmt.Println("Failed to convert to proto:", err)
		os.Exit(1)
	}
	site.CountryCode = "adblock"

	// 创建 GeoSiteList 并添加 site
	protoList := &router.GeoSiteList{
		Entry: []*router.GeoSite{site},
	}

	// 排序以保证可重复生成
	sort.SliceStable(protoList.Entry, func(i, j int) bool {
		return protoList.Entry[i].CountryCode < protoList.Entry[j].CountryCode
	})

	// 序列化为 proto 数据
	protoBytes, err := proto.Marshal(protoList)
	if err != nil {
		fmt.Println("Failed to marshal proto:", err)
		os.Exit(1)
	}

	// 写入 adblock.dat 文件
	if err := os.WriteFile(outputFile, protoBytes, 0644); err != nil {
		fmt.Println("Failed to write output file:", err)
		os.Exit(1)
	} else {
		fmt.Println(outputFile, "has been generated successfully.")
	}
}
