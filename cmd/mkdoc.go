package main

import (
	"docspace"
	"docspace/generators/markdown"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"strings"
)

import (
	_ "docspace/scanners/funcdoc"
	_ "github.com/thewinds/gqlcorego"
)

var (
	scannerName = kingpin.Flag("scanner", "which api scanner to use,eg. gql-corego").Required().Short('s').String()
	tag         = kingpin.Flag("tag", "which tag to filter,eg. v1").Short('t').String()
	pkg         = kingpin.Arg("pkg", "which package to scan").Required().String()
	output      = kingpin.Arg("out", "which file to output").String()
)

func main() {
	kingpin.Parse()
	scanners := docspace.GetScanners()
	scanner := scanners[*scannerName]
	if scanner == nil {
		fmt.Printf("error: scanner \"%s\" is not found,you can choose scanner below :\n", *scannerName)
		for name := range scanners {
			fmt.Printf("    %s\n", name)
		}
		return
	}
	goPaths := docspace.GetGOPaths()
	pkgExist := false
	for _, gopath := range goPaths {
		if _, err := os.Stat(filepath.Join(gopath, "src", *pkg)); err == nil {
			pkgExist = true
		}
	}
	if !pkgExist {
		fmt.Printf("error: package \"%s\" is not found in any of:\n", *pkg)
		for _, gopath := range goPaths {
			fmt.Println("  ", filepath.Join(gopath, "src", *pkg))
		}
		return
	}
	fmt.Printf("🔎  scan doc annotations (use %s)\n", scanner.GetName())
	annotations, err := scanner.ScanAnnotations(*pkg)
	if err != nil {
		fmt.Printf("error: scan annotations %v\n", err)
	}

	apis := make([]*docspace.API, 0, len(annotations))
	for k, a := range annotations {
		fmt.Printf("\r🔥  parse annotation to api [%d/%d]", k+1, len(annotations))
		api, err := a.ParseToAPI()
		if err != nil {
			fmt.Printf("error: annotation can not be parse,%v\n", err)
			return
		}
		apis = append(apis, api)
	}
	fmt.Printf("\n")

	// match tags
	matchTagAPIs := make([]*docspace.API, 0)
	tagsMap := map[string]bool{}
	allTags := make([]string, 0)

	if *tag != "" {
		for _, api := range apis {
			for _, t := range api.Tags {
				if _, exist := tagsMap[t]; !exist {
					tagsMap[t] = true
					allTags = append(allTags, t)
				}
				if t == *tag {
					matchTagAPIs = append(matchTagAPIs, api)
					break
				}
			}
		}
	} else {
		matchTagAPIs = apis
	}

	if len(matchTagAPIs) == 0 {
		fmt.Printf("👽  no tag is matched,all tags:\n")
		for _, t := range allTags {
			fmt.Printf("    %s\n", t)
		}
		return
	}

	if len(*tag) != 0 {
		fmt.Printf("👽  tag '%s' match %d api\n", *tag, len(matchTagAPIs))
	} else {
		fmt.Printf("👽  %d api is matched \n", len(matchTagAPIs))
	}
	// generate markdown

	markdownBuilder := strings.Builder{}
	markdownBuilder.WriteString(fmt.Sprintf("# %s API\n", *tag))
	markdownBuilder.WriteString("[TOC]\n")
	for k, api := range matchTagAPIs {
		fmt.Printf("\r🔥  building api '%s' [%d/%d]          ", api.Name, k+1, len(matchTagAPIs))
		err := api.Build()
		if err != nil {
			fmt.Printf("error: build api %s,%v\n", api.Name, err)
			return
		}
		output, _ := markdown.NewGenerator().Source(api).Gen()
		markdownBuilder.WriteString(output)
	}
	fmt.Println()

	outputFileName := "api_doc.md"
	if *output != "" {
		outputFileName = *output
	}
	fmt.Printf("📖  write api doc to '%s'\n", outputFileName)
	file, err := os.OpenFile(outputFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("error: witre file ,%v\n", err)
		return
	}
	_, err = file.WriteString(markdownBuilder.String())
	if err != nil {
		fmt.Printf("error: witre file ,%v\n", err)
		return
	}
	file.Close()
	fmt.Printf("🍺  done!\n")
}
