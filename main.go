package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"v2ray.com/core/app/router"
)

const (
	inputFileName  = "adblock_reject_domain_geosite.txt" // 输入文件名
	outputFileName = "adblock.dat"                       // 输出文件名
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
	Name  string
	Entry []Entry
}

func (l *ParsedList) toProto() (*router.GeoSite, error) {
	site := &router.GeoSite{
		CountryCode: l.Name,
	}
	for _, entry := range l.Entry {
		if entry.Type == "domain" { // 仅保留 "domain" 类型的域名条目
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Domain,
				Value:     entry.Value,
				Attribute: entry.Attrs,
			})
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

func main() {
	// 获取当前目录下的 adblock_reject_domain_geosite.txt 文件
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory:", err)
		return
	}
	filePath := filepath.Join(dir, inputFileName)

	// 加载文件内容
	list, err := Load(filePath)
	if err != nil {
		fmt.Println("Failed to load list:", err)
		return
	}

	// 解析并生成 ProtoBuf
	pl := &ParsedList{
		Name:  list.Name,
		Entry: list.Entry,
	}

	site, err := pl.toProto()
	if err != nil {
		fmt.Println("Failed to convert to Proto:", err)
		return
	}

	protoList := &router.GeoSiteList{
		Entry: []*router.GeoSite{site},
	}

	protoBytes, err := proto.Marshal(protoList)
	if err != nil {
		fmt.Println("Failed to marshal Proto:", err)
		return
	}

	// 将 ProtoBuf 数据写入 adblock.dat 文件
	if err := os.WriteFile(outputFileName, protoBytes, 0777); err != nil {
		fmt.Println("Failed to write file:", err)
	} else {
		fmt.Println(outputFileName, "has been generated successfully.")
	}
}
