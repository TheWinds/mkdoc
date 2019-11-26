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
	scannerName = kingpin.Flag("scanner", "which api scanner to use,eg. funcdoc").Required().Short('s').String()
	tag         = kingpin.Flag("tag", "which tag to filter,eg. v1").Short('t').String()
	mod         = kingpin.Flag("mod", "use go mod").Short('m').Bool()
	pkg         = kingpin.Arg("pkg", "which package to scan").Required().String()
	output      = kingpin.Arg("out", "which file to output").String()
)

func checkScanner() []docspace.APIScanner {
	var okScanners []docspace.APIScanner
	scanners := docspace.GetScanners()
	if scannerName == nil {
		return nil
	}
	for _, name := range strings.Split(*scannerName, ",") {
		if scanners[name] == nil {
			fmt.Printf("error: scanner \"%s\" is not found,you can choose scanner below :\n", name)
			for n := range scanners {
				fmt.Printf("    %s\n", n)
			}
			return nil
		}
		okScanners = append(okScanners, scanners[name])
	}
	return okScanners
}

func main() {
	kingpin.Parse()
	scanners := checkScanner()
	if scanners == nil {
		return
	}
	if *mod {
		docspace.UseGOModule = true
	}
	if !docspace.UseGOModule {
		// check if package exist
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
	} else {
		path := *pkg
		if !filepath.IsAbs(path) {
			wd, err := os.Getwd()
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			path = filepath.Join(wd, path)
		}
		if _, err := os.Stat(path); err != nil {
			fmt.Printf("no such file or directory: %s\n", path)
			return
		}
	}

	var apis []*docspace.API

	for _, scanner := range scanners {
		fmt.Printf("🔎  scan doc annotations (use %s)\n", scanner.GetName())
		annotations, err := scanner.ScanAnnotations(*pkg)
		if err != nil {
			fmt.Printf("error: scan annotations %v\n", err)
		}
		for k, a := range annotations {
			fmt.Printf("\r🔥  parse annotation to api [%d/%d]", k+1, len(annotations))
			api, err := a.ParseToAPI()
			if err != nil {
				fmt.Printf("\n❌  annotation can not be parse\n%v\n------\nAnnotation:%s\n------\n", err, a)
				return
			}
			apis = append(apis, api)
		}
		fmt.Printf("\n")
	}

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
			fmt.Printf("\n❌  build api %s\n%v\n------\nAnnotation:\n%s\n------\n", api.Name, err, api.Annotation)
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
