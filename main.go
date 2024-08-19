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

type Entry struct {
	Type  string
	Value string
	Attrs []*router.Domain_Attribute
}

type List struct {
	Name  string
	Entry []Entry
}

func (l *List) toProto() (*router.GeoSite, error) {
	site := &router.GeoSite{
		CountryCode: "adblock", // 强制设置为 "adblock"
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

	return entry, nil
}

func Load(path string) (*List, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list := &List{
		Name: "adblock", // 强制设置为 "adblock"
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
	inputFilePath := "adblock_reject_domain_geosite.txt"
	outputFilePath := "adblock.dat"

	list, err := Load(inputFilePath)
	if err != nil {
		fmt.Println("Failed to load file:", err)
		return
	}

	site, err := list.toProto()
	if err != nil {
		fmt.Println("Failed to convert to proto:", err)
		return
	}

	protoBytes, err := proto.Marshal(site)
	if err != nil {
		fmt.Println("Failed to marshal proto:", err)
		return
	}
	if err := os.WriteFile(outputFilePath, protoBytes, 0777); err != nil {
		fmt.Println("Failed to write file:", err)
	} else {
		fmt.Println("adblock.dat has been generated successfully.")
	}
}
