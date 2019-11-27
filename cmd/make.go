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

func makeDoc(ctx *kingpin.ParseContext) error {
	project, err := readProjectConfig()
	if err != nil {
		return fmt.Errorf("fail to read config file: %v", err)
	}

	scanners := checkScanner(project)
	if scanners == nil {
		return nil
	}

	if !project.UseGOModule {
		// check if package exist
		goPaths := docspace.GetGOPaths()
		pkgExist := false
		for _, gopath := range goPaths {
			if _, err := os.Stat(filepath.Join(gopath, "src", project.BasePackage)); err == nil {
				pkgExist = true
			}
		}
		if !pkgExist {
			fmt.Printf("error: package \"%s\" is not found in any of:\n", project.BasePackage)
			for _, gopath := range goPaths {
				fmt.Println("  ", filepath.Join(gopath, "src", project.BasePackage))
			}
			return nil
		}
	} else {
		path := project.BasePackage
		if !filepath.IsAbs(path) {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			path = filepath.Join(wd, path)
		}
		if _, err := os.Stat(path); err != nil {
			fmt.Printf("no such file or directory: %s\n", path)
			return nil
		}
	}

	var apis []*docspace.API

	for _, scanner := range scanners {
		fmt.Printf("🔎  scan doc annotations (use %s)\n", scanner.GetName())
		annotations, err := scanner.ScanAnnotations(*project)
		if err != nil {
			fmt.Printf("error: scan annotations %v\n", err)
		}
		for k, a := range annotations {
			fmt.Printf("\r🔥  parse annotation to api [%d/%d]", k+1, len(annotations))
			api, err := a.ParseToAPI()
			if err != nil {
				fmt.Printf("\n❌  annotation can not be parse\n%v\n------\nAnnotation:%s\n------\n", err, a)
				return nil
			}
			api.Project = project
			apis = append(apis, api)
		}
		fmt.Printf("\n")
	}

	// match tags
	matchTagAPIs := make([]*docspace.API, 0)
	tagsMap := map[string]bool{}
	allTags := make([]string, 0)

	tag := *makeDocTag
	if tag != "" {
		for _, api := range apis {
			for _, t := range api.Tags {
				if _, exist := tagsMap[t]; !exist {
					tagsMap[t] = true
					allTags = append(allTags, t)
				}
				if t == tag {
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
		return nil
	}

	if tag != "" {
		fmt.Printf("👽  tag '%s' match %d api\n", tag, len(matchTagAPIs))
	} else {
		fmt.Printf("👽  %d api is matched \n", len(matchTagAPIs))
	}
	// generate markdown

	markdownBuilder := strings.Builder{}
	markdownBuilder.WriteString(fmt.Sprintf("# %s API\n", tag))
	markdownBuilder.WriteString("[TOC]\n")
	for k, api := range matchTagAPIs {
		fmt.Printf("\r🔥  building api '%s' [%d/%d]          ", api.Name, k+1, len(matchTagAPIs))
		err := api.Build()
		if err != nil {
			fmt.Printf("\n❌  build api %s\n%v\n------\nAnnotation:\n%s\n------\n", api.Name, err, api.Annotation)
			return nil
		}
		output, _ := markdown.NewGenerator().Source(api).Gen()
		markdownBuilder.WriteString(output)
	}
	fmt.Println()
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	if tag == "" {
		tag = "mydoc"
	}
	outputFileName := fmt.Sprintf("%s.md", tag)
	out := *makeDocOut
	if out != "" {
		outputFileName = out
	}
	fmt.Printf("📖  write api doc to '%s'\n", outputFileName)
	file, err := os.OpenFile(filepath.Join(path, "docs", outputFileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("error: witre file ,%v\n", err)
		return nil
	}
	_, err = file.WriteString(markdownBuilder.String())
	if err != nil {
		fmt.Printf("error: witre file ,%v\n", err)
		return nil
	}
	file.Close()
	fmt.Printf("🍺  done!\n")
	return nil
}
