package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
)

var (
	datName      = flag.String("datname", "adblock.dat", "Name of the generated dat file")
	outputPath   = flag.String("outputpath", "./publish", "Output path to the generated files")
	exportLists  = flag.String("exportlists", "", "Lists to be exported in plaintext format, separated by ',' comma")
	excludeAttrs = flag.String("excludeattrs", "", "Exclude rules with certain attributes in certain lists, separated by ',' comma")
	toGFWList    = flag.String("togfwlist", "", "List to be exported in GFWList format")
)

func main() {
	flag.Parse()

	adblockFile := "./adblock.txt"
	listInfoMap := make(ListInfoMap)

	if err := listInfoMap.Marshal(adblockFile); err != nil {
		fmt.Println("Failed:", err)
		os.Exit(1)
	}

	if err := listInfoMap.FlattenAndGenUniqueDomainList(); err != nil {
		fmt.Println("Failed:", err)
		os.Exit(1)
	}

	excludeAttrsInFile := make(map[fileName]map[attribute]bool)
	if *excludeAttrs != "" {
		exFilenameAttrSlice := strings.Split(*excludeAttrs, ",")
		for _, exFilenameAttr := range exFilenameAttrSlice {
			exFilenameAttr = strings.TrimSpace(exFilenameAttr)
			exFilenameAttrMap := strings.Split(exFilenameAttr, "@")
			filename := fileName(strings.ToUpper(strings.TrimSpace(exFilenameAttrMap[0])))
			excludeAttrsInFile[filename] = make(map[attribute]bool)
			for _, attr := range exFilenameAttrMap[1:] {
				attr = strings.TrimSpace(attr)
				if len(attr) > 0 {
					excludeAttrsInFile[filename][attribute(attr)] = true
				}
			}
		}
	}

	var exportListsSlice []string
	if *exportLists != "" {
		tempSlice := strings.Split(*exportLists, ",")
		for _, exportList := range tempSlice {
			exportList = strings.TrimSpace(exportList)
			if len(exportList) > 0 {
				exportListsSlice = append(exportListsSlice, exportList)
			}
		}
	}

	if geositeList := listInfoMap.ToProto(excludeAttrsInFile); geositeList != nil {
		protoBytes, err := proto.Marshal(geositeList)
		if err != nil {
			fmt.Println("Failed:", err)
			os.Exit(1)
		}
		if err := os.MkdirAll(*outputPath, 0755); err != nil {
			fmt.Println("Failed:", err)
			os.Exit(1)
		}
		if err := os.WriteFile(filepath.Join(*outputPath, *datName), protoBytes, 0644); err != nil {
			fmt.Println("Failed:", err)
			os.Exit(1)
		} else {
			fmt.Printf("%s has been generated successfully in '%s'.\n", *datName, *outputPath)
		}
	}

	if filePlainTextBytesMap, err := listInfoMap.ToPlainText(exportListsSlice); err == nil {
		for filename, plaintextBytes := range filePlainTextBytesMap {
			filename += ".txt"
			if err := os.WriteFile(filepath.Join(*outputPath, filename), plaintextBytes, 0644); err != nil {
				fmt.Println("Failed:", err)
				os.Exit(1)
			} else {
				fmt.Printf("%s has been generated successfully in '%s'.\n", filename, *outputPath)
			}
		}
	} else {
		fmt.Println("Failed:", err)
		os.Exit(1)
	}

	if gfwlistBytes, err := listInfoMap.ToGFWList(*toGFWList); err == nil {
		if f, err := os.OpenFile(filepath.Join(*outputPath, "gfwlist.txt"), os.O_RDWR|os.O_CREATE, 0644); err != nil {
			fmt.Println("Failed:", err)
			os.Exit(1)
		} else {
			encoder := base64.NewEncoder(base64.StdEncoding, f)
			defer encoder.Close()
			if _, err := encoder.Write(gfwlistBytes); err != nil {
				fmt.Println("Failed:", err)
				os.Exit(1)
			}
			fmt.Printf("gfwlist.txt has been generated successfully in '%s'.\n", *outputPath)
		}
	} else {
		fmt.Println("Failed:", err)
		os.Exit(1)
	}
}
