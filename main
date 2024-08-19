package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"v2ray.com/core/app/router"
)

var (
	// 定义数据文件路径的命令行参数
	filePath = flag.String("filepath", "adblock_reject_domain_geosite.txt", "Path to your 'adblock_reject_domain_geosite.txt' file")
)

// Entry 代表一个域名条目
type Entry struct {
	Type  string
	Value string
	Attrs []*router.Domain_Attribute
}

// List 代表域名列表
type List struct {
	Name  string
	Entry []Entry
}

// ParsedList 代表解析后的域名列表
type ParsedList struct {
	Name      string
	Inclusion map[string]bool
	Entry     []Entry
}

// 将解析后的列表转换为 Protobuf 格式
func (l *ParsedList) toProto() (*router.GeoSite, error) {
	site := &router.GeoSite{
		CountryCode: l.Name,
	}
	for _, entry := range l.Entry {
		switch entry.Type {
		case "domain":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Domain,
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

// 移除行中的注释部分
func removeComment(line string) string {
	idx := strings.Index(line, "#")
	if idx == -1 {
		return line
	}
	return strings.TrimSpace(line[:idx])
}

// 解析域名
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

// 解析域名属性
func parseAttribute(attr string) (router.Domain_Attribute, error) {
	var attribute router.Domain_Attribute
	if len(attr) == 0 || attr[0] != '@' {
		return attribute, errors.New("invalid attribute: " + attr)
	}

	attr = attr[0:]
	parts := strings.Split(attr, "=")
	if len(parts) == 1 {
		attribute.Key = strings.ToLower(parts[0])
		attribute.TypedValue = &router.Domain_Attribute_BoolValue{BoolValue: true}
	} else {
		attribute.Key = strings.ToLower(parts[0])
		intv, err := strconv.Atoi(parts[1])
		if err != nil {
			return attribute, errors.New("invalid attribute: " + attr + ": " + err.Error())
		}
		attribute.TypedValue = &router.Domain_Attribute_IntValue{IntValue: int64(intv)}
	}
	return attribute, nil
}

// 解析条目
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
		entry.Attrs = append(entry.Attrs, &attr)
	}

	return entry, nil
}

// 加载域名列表
func Load(path string) (*List, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list := &List{
		Name: "ADBLOCK",
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

func main() {
	flag.Parse()

	// 读取指定的文件路径
	fmt.Println("Loading domain list from:", *filePath)
	list, err := Load(*filePath)
	if err != nil {
		fmt.Println("Failed to load domain list:", err)
		return
	}

	// 解析域名列表并转换为 Protobuf 格式
	pl, err := ParseList(list, nil)
	if err != nil {
		fmt.Println("Failed to parse domain list:", err)
		return
	}

	site, err := pl.toProto()
	if err != nil {
		fmt.Println("Failed to convert to proto:", err)
		return
	}

	// 生成并保存二进制文件
	protoList := &router.GeoSiteList{Entry: []*router.GeoSite{site}}
	data, err := proto.Marshal(protoList)
	if err != nil {
		fmt.Println("Error marshaling proto:", err)
		return
	}

	output := "adblock_geosite.dat"
	if err := ioutil.WriteFile(output, data, 0644); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("GeoSite file generated:", output)
}
