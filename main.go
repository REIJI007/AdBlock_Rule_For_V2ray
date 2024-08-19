package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"v2ray.com/core/app/router"
)

// Entry 结构体表示一个域名条目，包含域名类型、值和属性。
type Entry struct {
	Type  string                    // 域名类型，可以是 "domain", "regexp", "keyword", "full"。
	Value string                    // 域名值。
	Attrs []*router.Domain_Attribute // 域名的属性。
}

// List 结构体表示一个域名列表，包含列表名称和多个域名条目。
type List struct {
	Name  string  // 列表名称。
	Entry []Entry // 域名条目列表。
}

// toProto 将 List 转换为 protobuf 格式的 GeoSite 对象。
func (l *List) toProto() (*router.GeoSite, error) {
	site := &router.GeoSite{
		CountryCode: "adblock", // 强制设置为 "adblock"。
	}
	for _, entry := range l.Entry {
		switch entry.Type {
		case "domain":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Domain, // 域名类型为普通域名。
				Value:     entry.Value,          // 设置域名值。
				Attribute: entry.Attrs,          // 设置域名属性。
			})
		case "regexp":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Regex,  // 域名类型为正则表达式。
				Value:     entry.Value,          // 设置正则表达式值。
				Attribute: entry.Attrs,          // 设置域名属性。
			})
		case "keyword":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Plain,  // 域名类型为关键字。
				Value:     entry.Value,          // 设置关键字值。
				Attribute: entry.Attrs,          // 设置域名属性。
			})
		case "full":
			site.Domain = append(site.Domain, &router.Domain{
				Type:      router.Domain_Full,   // 域名类型为完整域名。
				Value:     entry.Value,          // 设置完整域名值。
				Attribute: entry.Attrs,          // 设置域名属性。
			})
		default:
			return nil, errors.New("unknown domain type: " + entry.Type) // 处理未知的域名类型错误。
		}
	}
	return site, nil
}

// removeComment 移除行中的注释（以 "#" 开头的部分）。
func removeComment(line string) string {
	idx := strings.Index(line, "#")
	if idx == -1 {
		return line
	}
	return strings.TrimSpace(line[:idx])
}

// parseDomain 解析域名字符串，将其转换为 Entry 结构体。
func parseDomain(domain string, entry *Entry) error {
	kv := strings.Split(domain, ":")
	if len(kv) == 1 {
		entry.Type = "domain"                  // 如果没有 ":"，则默认类型为 "domain"。
		entry.Value = strings.ToLower(kv[0])   // 设置域名值为小写。
		return nil
	}

	if len(kv) == 2 {
		entry.Type = strings.ToLower(kv[0])    // 解析域名类型。
		entry.Value = strings.ToLower(kv[1])   // 解析域名值。
		return nil
	}

	return errors.New("Invalid format: " + domain) // 处理无效格式错误。
}

// parseEntry 解析一行文本，将其转换为 Entry 结构体。
func parseEntry(line string) (Entry, error) {
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")

	var entry Entry
	if len(parts) == 0 {
		return entry, errors.New("empty entry") // 处理空条目错误。
	}

	if err := parseDomain(parts[0], &entry); err != nil {
		return entry, err // 解析域名并处理错误。
	}

	return entry, nil
}

// Load 从指定路径加载域名列表文件，并解析为 List 结构体。
func Load(path string) (*List, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err // 处理文件打开错误。
	}
	defer file.Close()

	list := &List{
		Name: "adblock", // 强制设置为 "adblock"。
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // 去除行首尾空白字符。
		line = removeComment(line)                // 去除行内注释。
		if len(line) == 0 {
			continue // 跳过空行。
		}
		entry, err := parseEntry(line)
		if err != nil {
			return nil, err // 处理条目解析错误。
		}
		list.Entry = append(list.Entry, entry) // 添加解析后的条目到列表。
	}

	return list, nil
}

// main 函数是程序的入口点。
func main() {
	inputFilePath := "adblock_reject_domain_geosite.txt" // 输入文件路径。
	outputFilePath := "adblock.dat"                      // 输出文件路径。

	list, err := Load(inputFilePath)
	if err != nil {
		fmt.Println("Failed to load file:", err) // 处理文件加载错误。
		return
	}

	site, err := list.toProto()
	if err != nil {
		fmt.Println("Failed to convert to proto:", err) // 处理转换为 protobuf 错误。
		return
	}

	protoBytes, err := proto.Marshal(site)
	if err != nil {
		fmt.Println("Failed to marshal proto:", err) // 处理序列化错误。
		return
	}
	if err := os.WriteFile(outputFilePath, protoBytes, 0777); err != nil {
		fmt.Println("Failed to write file:", err) // 处理文件写入错误。
	} else {
		fmt.Println("adblock.dat has been generated successfully.") // 成功生成文件。
	}
}
