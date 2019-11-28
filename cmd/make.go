// command: mkdoc make
package main

import (
	"docspace"
	"docspace/generators/markdown"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

import (
	_ "docspace/scanners/funcdoc"
	_ "github.com/thewinds/gqlcorego"
)

func readProjectConfig() (*docspace.Project, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	confFilePath := filepath.Join(path, configFileName)
	b, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}
	project := new(docspace.Project)
	err = yaml.Unmarshal(b, project)
	return project, err
}

func checkScanner(project *docspace.Project) []docspace.APIScanner {
	var okScanners []docspace.APIScanner
	scanners := docspace.GetScanners()
	if len(project.Scanner) == 0 {
		fmt.Printf("error: please use at least one scanner \n")
		return nil
	}

	for _, name := range project.Scanner {
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

func checkConfig(project *docspace.Project) error {
	// check scanners
	scanners := checkScanner(project)
	if scanners == nil {
		return fmt.Errorf("please config scanner in conf.yaml")
	}
	project.OnScanners = scanners

	// check if the pkg to scan is exist
	if project.BasePackage == "" {
		return fmt.Errorf("please config a pkg to scan in conf.yaml")
	}

	if project.UseGOModule {
		path := project.BasePackage
		if !filepath.IsAbs(path) {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			path = filepath.Join(wd, path)
		}
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("no such file or directory: %s\n", path)
		}
	} else {
		goPaths := docspace.GetGOPaths()
		pkgExist := false
		for _, gopath := range goPaths {
			if _, err := os.Stat(filepath.Join(gopath, "src", project.BasePackage)); err == nil {
				pkgExist = true
			}
		}
		if !pkgExist {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("error: package \"%s\" is not found in any of:\n", project.BasePackage))
			for _, gopath := range goPaths {
				sb.WriteString(fmt.Sprintln("  ", filepath.Join(gopath, "src", project.BasePackage)))
			}
			return fmt.Errorf("%s", sb.String())
		}
	}
	return nil
}

func scanAPIs(project *docspace.Project) ([]*docspace.API, error) {
	var apis []*docspace.API
	for _, scanner := range project.OnScanners {
		fmt.Printf("🔎  scan doc annotations (use %s)\n", scanner.GetName())
		annotations, err := scanner.ScanAnnotations(*project)
		if err != nil {
			return nil, fmt.Errorf("scan annotations %v\n", err)
		}
		for k, a := range annotations {
			fmt.Printf("\r🔥  parse annotation to api [%d/%d]", k+1, len(annotations))
			api, err := a.ParseToAPI()
			if err != nil {
				fmt.Println()
				return nil, fmt.Errorf("annotation can not be parse\n%v\n------\nAnnotation:%s\n------\n", err, a)
			}
			api.Project = project
			apis = append(apis, api)
		}
		fmt.Printf("\n")
	}
	return apis, nil
}

func getAllTags(apis []*docspace.API) []string {
	tagsMap := make(map[string]bool)
	for _, api := range apis {
		for _, t := range api.Tags {
			tagsMap[t] = true
		}
	}
	var tags []string
	for tag := range tagsMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func filterAPIByTag(apis []*docspace.API, tag string) []*docspace.API {
	var matched []*docspace.API

	if tag == "" {
		return apis
	}

	for _, api := range apis {
		for _, t := range api.Tags {
			if t == tag {
				matched = append(matched, api)
				break
			}
		}
	}
	return matched
}

func buildAPI(apis []*docspace.API) error {
	for k, api := range apis {
		fmt.Printf("\r🔥  building api '%s' [%d/%d]          ", api.Name, k+1, len(apis))
		err := api.Build()
		if err != nil {
			fmt.Println()
			return fmt.Errorf("build api %s\n%v\n------\nAnnotation:\n%s\n------\n", api.Name, err, api.Annotation)
		}
	}
	fmt.Println()
	return nil
}

func makeDoc(ctx *kingpin.ParseContext) error {
	project, err := readProjectConfig()
	if err != nil {
		return showErr("fail to read config file: %v", err)
	}

	if err := checkConfig(project); err != nil {
		return showErr("%v", err)
	}

	apis, err := scanAPIs(project)
	if err != nil {
		return showErr("%v", err)
	}

	tag := *makeDocTag

	matchedAPIs := filterAPIByTag(apis, tag)

	if len(matchedAPIs) == 0 {
		fmt.Printf("👽  no tag is matched,all tags:\n")
		for _, t := range getAllTags(apis) {
			fmt.Printf("    %s\n", t)
		}
		return nil
	}

	if tag != "" {
		fmt.Printf("👽  tag '%s' match %d api\n", tag, len(matchedAPIs))
	} else {
		fmt.Printf("👽  %d api is matched \n", len(matchedAPIs))
	}

	if err := buildAPI(matchedAPIs); err != nil {
		return showErr("%v", err)
	}

	if err := genMarkdown(project, matchedAPIs, tag); err != nil {
		return showErr("%v", err)
	}
	fmt.Printf("🍺  done!\n")
	return nil
}

func genMarkdown(project *docspace.Project, apis []*docspace.API, tag string) error {
	markdownBuilder := strings.Builder{}
	if tag == "" {
		tag = "all"
	}
	header := `
# %s

> %s 

##  Summary

| 📖 **Tag**     | %s |
| ------------- | ------ |
| 🔮 **API Num** | %s   |

[TOC]

# API List
`
	markdownBuilder.WriteString(fmt.Sprintf(header,
		project.Name,
		project.Description,
		"`"+tag+"`",
		fmt.Sprintf("`%d`", len(apis))))

	for _, api := range apis {
		output, _ := markdown.NewGenerator().Source(api).Gen()
		markdownBuilder.WriteString(output)
	}
	fmt.Println()

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	outputFileName := fmt.Sprintf("%s.md", tag)
	out := *makeDocOut
	if out != "" {
		outputFileName = out
	}
	fmt.Printf("📖  write api doc to './docs/md/%s'\n", outputFileName)
	mdPath := filepath.Join(path, "docs/md")
	if _, err = os.Stat(mdPath); err != nil {
		err = os.Mkdir(mdPath, 0755)
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile(filepath.Join(mdPath, outputFileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("witre file ,%v\n", err)
	}
	defer file.Close()
	_, err = file.WriteString(markdownBuilder.String())
	if err != nil {
		return fmt.Errorf("witre file ,%v\n", err)
	}
	return nil
}
