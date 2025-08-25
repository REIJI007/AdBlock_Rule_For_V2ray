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

// 定义命令行标志
var (
	dataPath    = flag.String("datapath", "./", "指定数据文件的目录为仓库根目录") // 修改为仓库根目录
	outputName  = flag.String("outputname", "adblock.dat", "生成的dat文件的名称") // 修改输出文件名为 adblock.dat
	outputDir   = flag.String("outputdir", "./", "生成文件的存放目录")
	exportLists = flag.String("exportlists", "", "要导出的纯文本格式列表，多个列表用逗号分隔")
)

// 定义Entry结构，表示域名条目
type Entry struct {
	Type  string                      // 域名类型 (domain, regexp, keyword, full)
	Value string                      // 域名值
	Attrs []*router.Domain_Attribute   // 域名的附加属性
}

// 定义List结构，表示包含多个Entry的列表
type List struct {
	Name  string   // 列表名称
	Entry []Entry  // 列表中的条目
}

// 定义ParsedList结构，包含解析后的域名条目
type ParsedList struct {
	Name      string              // 列表名称
	Inclusion map[string]bool      // 已包含的列表
	Entry     []Entry             // 解析后的条目
}

// 将ParsedList转换为纯文本并写入文件
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
		// 以 "type:domain.tld:@attr1,@attr2" 的格式保存条目
		entryBytes = append(entryBytes, []byte(entry.Type+":"+entry.Value+attrString+"\n")...)
	}
	if err := os.WriteFile(filepath.Join(*outputDir, listName+".txt"), entryBytes, 0644); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

// 将ParsedList转换为Proto格式
func (l *ParsedList) toProto() (*router.GeoSite, error) {
	site := &router.GeoSite{
		CountryCode: l.Name, // 使用列表名称作为国家代码
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
			return nil, errors.New("未知的域名类型: " + entry.Type)
		}
	}
	return site, nil
}

// 导出指定列表为纯文本格式
func exportPlainTextList(list []string, refName string, pl *ParsedList) {
	for _, listName := range list {
		if strings.EqualFold(refName, listName) {
			if err := pl.toPlainText(strings.ToLower(refName)); err != nil {
				fmt.Println("导出失败: ", err)
				continue
			}
			fmt.Printf("'%s' 已成功生成。\n", listName)
		}
	}
}

// 移除行内注释
func removeComment(line string) string {
	idx := strings.Index(line, "#")
	if idx == -1 {
		return line
	}
	return strings.TrimSpace(line[:idx])
}

// 解析域名条目
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

	return errors.New("无效的格式: " + domain)
}

// 解析属性
func parseAttribute(attr string) (*router.Domain_Attribute, error) {
	var attribute router.Domain_Attribute
	if len(attr) == 0 || attr[0] != '@' {
		return &attribute, errors.New("无效的属性: " + attr)
	}

	// 去除属性前缀 `@`
	attr = attr[1:]
	parts := strings.Split(attr, "=")
	if len(parts) == 1 {
		attribute.Key = strings.ToLower(parts[0])
		attribute.TypedValue = &router.Domain_Attribute_BoolValue{BoolValue: true}
	} else {
		attribute.Key = strings.ToLower(parts[0])
		intv, err := strconv.Atoi(parts[1])
		if err != nil {
			return &attribute, errors.New("无效的属性: " + attr + ": " + err.Error())
		}
		attribute.TypedValue = &router.Domain_Attribute_IntValue{IntValue: int64(intv)}
	}
	return &attribute, nil
}

// 解析域名条目
func parseEntry(line string) (Entry, error) {
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")

	var entry Entry
	if len(parts) == 0 {
		return entry, errors.New("空条目")
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

// 加载指定路径的列表文件
func Load(path string) (*List, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list := &List{
		Name: "ADBLOCK", // 修改列表标签为ADBLOCK
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

// 判断条目的属性是否匹配
func isMatchAttr(Attrs []*router.Domain_Attribute, includeKey string) bool {
	isMatch := false
	mustMatch := true
	matchName := includeKey
	if strings.HasPrefix(includeKey, "!") {
		isMatch = true
		mustMatch = false
		matchName = strings.TrimLeft(includeKey, "!")
	}

	for _, Attr := range Attrs {
		attrName := Attr.Key
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

// 创建匹配属性的条目列表
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

// 解析列表，并递归处理包含的其他列表
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
				InclusionName := strings.ToUpper(entry.Value) // 使用InclusionName代替refName
				if strings.HasPrefix(InclusionName, "ATTR@") {
					attr := &router.Domain_Attribute{
						Key: strings.ToLower(InclusionName[5:]),
					}
					for _, refList := range ref {
						attrEntrys := createIncludeAttrEntrys(refList, attr)
						if len(attrEntrys) != 0 {
							newEntryList = append(newEntryList, attrEntrys...)
						}
					}
				} else {
					if pl.Inclusion[InclusionName] {
						continue
					}
					pl.Inclusion[InclusionName] = true
					refList := ref[InclusionName]
					if refList == nil {
						return nil, errors.New(entry.Value + " 找不到。")
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

// 主函数
func main() {
	flag.Parse()

	// 设定adblock.txt为读取文件
	filePath := filepath.Join(".", "adblock.txt")
	fmt.Println("使用域名列表文件: ", filePath)

	ref := make(map[string]*List)
	list, err := Load(filePath)
	if err != nil {
		fmt.Println("加载失败: ", err)
		os.Exit(1)
	}
	ref[list.Name] = list

	// 如果输出目录不存在，创建输出目录
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(*outputDir, 0755); mkErr != nil {
			fmt.Println("创建目录失败: ", mkErr)
			os.Exit(1)
		}
	}

	protoList := new(router.GeoSiteList)
	pl, err := ParseList(list, ref)
	if err != nil {
		fmt.Println("解析失败: ", err)
		os.Exit(1)
	}
	site, err := pl.toProto()
	if err != nil {
		fmt.Println("转换失败: ", err)
		os.Exit(1)
	}
	protoList.Entry = append(protoList.Entry, site)

	// 对protoList进行排序，确保输出的一致性
	sort.SliceStable(protoList.Entry, func(i, j int) bool {
		return protoList.Entry[i].CountryCode < protoList.Entry[j].CountryCode
	})

	protoBytes, err := proto.Marshal(protoList)
	if err != nil {
		fmt.Println("生成失败:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(*outputDir, *outputName), protoBytes, 0644); err != nil {
		fmt.Println("写入文件失败: ", err)
		os.Exit(1)
	} else {
		fmt.Println(*outputName, "已成功生成。")
	}
}
