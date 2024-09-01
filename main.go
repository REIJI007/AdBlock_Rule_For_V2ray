package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		// 将所有域名条目类型设置为 domain
		site.Domain = append(site.Domain, &router.Domain{
			Type:      router.Domain_Domain,
			Value:     entry.Value,
			Attribute: entry.Attrs,
		})
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
	entry.Type = "domain" // 强制所有条目类型为 domain
	entry.Value = strings.ToLower(domain)
	return nil
}

func parseEntry(line string) (Entry, error) {
	line = strings.TrimSpace(line)
	var entry Entry
	if len(line) == 0 {
		return entry, errors.New("empty entry")
	}

	if err := parseDomain(line, &entry); err != nil {
		return entry, err
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
		Name: "adblock", // 统一名称为 adblock
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
	// 获取当前目录下的adblock_reject_domain_geosite.txt文件
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

	// 解析并生成ProtoBuf
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

	// 将ProtoBuf数据写入adblock.dat文件
	if err := os.WriteFile(outputFileName, protoBytes, 0777); err != nil {
		fmt.Println("Failed to write file:", err)
	} else {
		fmt.Println(outputFileName, "has been generated successfully.")
	}
}
