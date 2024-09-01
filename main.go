package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
	"github.com/your/repo/protobuf/domainpb" // 假设你有一个用于序列化的 Protobuf package
)

var (
	inputFile   = flag.String("inputfile", filepath.Join("./", "adblock.txt"), "Path to the adblock.txt file")
	outputFile  = flag.String("outputfile", filepath.Join("./", "adblock.dat"), "Path to the generated adblock.dat file")
)

func main() {
	flag.Parse()

	// 读取 adblock.txt 文件
	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Println("Failed to open input file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// 初始化 DomainList
	var domainList domainpb.DomainList

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" && !strings.HasPrefix(domain, "#") {
			// 创建新的 Domain 对象并设置类型和类别
			d := &domainpb.Domain{
				Type: domainpb.Domain_ROUTER,
				Value: domain,
				Category: "adblock",
			}
			domainList.Domains = append(domainList.Domains, d)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Failed to scan input file:", err)
		os.Exit(1)
	}

	// 序列化到 .dat 文件
	protoBytes, err := proto.Marshal(&domainList)
	if err != nil {
		fmt.Println("Failed to marshal domain list:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, protoBytes, 0644); err != nil {
		fmt.Println("Failed to write output file:", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s.\n", *outputFile)
}
