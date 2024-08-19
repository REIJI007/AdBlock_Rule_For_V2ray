package main

import (
	"bufio"                     // 用于逐行读取文件
	"errors"                    // 用于错误处理
	"fmt"                       // 用于格式化输出
	"os"                        // 用于操作系统功能，如文件
	"path/filepath"             // 用于处理文件路径
	"strings"                   // 用于字符串操作

	"github.com/golang/protobuf/proto" // 用于处理Protobuf数据格式
	"v2ray.com/core/app/router"        // V2Ray的路由模块，包含域名解析等功能
)

// 定义条目结构体，表示域名列表中的一条记录
type Entry struct {
	Type  string                       // 域名类型（如domain、regexp等）
	Value string                       // 域名值
	Attrs []*router.Domain_Attribute   // 域名属性（如是否为广告域名等）
}

// 定义列表结构体，表示一个域名列表
type List struct {
	Name  string   // 列表名称
	Entry []Entry  // 列表中的所有条目
}

// 将解析后的列表转换为Protobuf格式的GeoSite对象
func (l *List) toProto() (*router.GeoSite, error) {
	site := &router.GeoSite{
		CountryCode: l.Name,  // 使用列表名称作为国家代码
	}
	for _, entry := range l.Entry {  // 遍历所有条目
		switch entry.Type {  // 根据条目类型创建对应的Protobuf对象
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

// 移除行中的注释部分（以#开头的部分）
func removeComment(line string) string {
	idx := strings.Index(line, "#")
	if idx == -1 {
		return line
	}
	return strings.TrimSpace(line[:idx])
}

// 解析域名条目，确定其类型和值
func parseDomain(domain string, entry *Entry) error {
	kv := strings.Split(domain, ":")
	if len(kv) == 1 {  // 如果没有冒号，默认为普通域名类型
		entry.Type = "domain"
		entry.Value = strings.ToLower(kv[0])
		return nil
	}

	if len(kv) == 2 {  // 含有冒号的情况，解析为指定类型的域名
		entry.Type = strings.ToLower(kv[0])
		entry.Value = strings.ToLower(kv[1])
		return nil
	}

	return errors.New("Invalid format: " + domain)
}

// 解析一行条目，生成Entry对象
func parseEntry(line string) (Entry, error) {
	line = strings.TrimSpace(line)  // 去除行首尾空白
	parts := strings.Split(line, " ")

	var entry Entry
	if len(parts) == 0 {  // 如果行为空，返回错误
		return entry, errors.New("empty entry")
	}

	if err := parseDomain(parts[0], &entry); err != nil {  // 解析域名部分
		return entry, err
	}

	return entry, nil
}

// 加载文件，解析成List结构体
func Load(path string) (*List, error) {
	file, err := os.Open(path)  // 打开文件
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list := &List{
		Name: strings.ToUpper(filepath.Base(path)),  // 使用文件名作为列表名称
	}
	scanner := bufio.NewScanner(file)  // 创建扫描器逐行读取文件
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())  // 去除每行的空白字符
		line = removeComment(line)  // 去除注释部分
		if len(line) == 0 {  // 跳过空行
			continue
		}
		entry, err := parseEntry(line)  // 解析条目
		if err != nil {
			return nil, err
		}
		list.Entry = append(list.Entry, entry)  // 添加到列表中
	}

	return list, nil
}

func main() {
	// 目标文件路径，假设当前工作目录为仓库根目录
	inputFilePath := "adblock_reject_domain_geosite.txt"
	outputFilePath := "adblock_geosite.dat"

	// 加载并解析输入文件
	list, err := Load(inputFilePath)
	if err != nil {
		fmt.Println("Failed to load file:", err)
		return
	}

	// 将解析后的列表转换为Protobuf格式
	site, err := list.toProto()
	if err != nil {
		fmt.Println("Failed to convert to proto:", err)
		return
	}

	// 将生成的ProtoBuf数据写入输出文件
	protoBytes, err := proto.Marshal(site)
	if err != nil {
		fmt.Println("Failed to marshal proto:", err)
		return
	}
	if err := ioutil.WriteFile(outputFilePath, protoBytes, 0777); err != nil {
		fmt.Println("Failed to write file:", err)
	} else {
		fmt.Println("adblock_geosite.dat has been generated successfully.")
	}
}
