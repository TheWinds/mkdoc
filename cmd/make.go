// command: mkdoc make
package main

import (
	"docspace"
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
	_ "docspace/generators/markdown"
	_ "docspace/scanners/funcdoc"
	_ "github.com/thewinds/gqlcorego"
)

func readConfig() (*docspace.Config, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	confFilePath := filepath.Join(path, configFileName)
	b, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}
	config := new(docspace.Config)
	err = yaml.Unmarshal(b, config)
	return config, err
}

func checkScanner(project *docspace.Project) bool {
	var okScanners []docspace.APIScanner
	scanners := docspace.GetScanners()
	if len(project.Config.Scanner) == 0 {
		showWarn("please use at least one scanner \n")
		return false
	}

	for _, name := range project.Config.Scanner {
		if scanners[name] == nil {
			showErr("scanner \"%s\" is not found,you can choose scanner below :\n", name)
			for n := range scanners {
				fmt.Printf("    %s\n", n)
			}
			return false
		}
		okScanners = append(okScanners, scanners[name])
	}
	project.Scanners = okScanners
	return true
}

func checkGenerator(project *docspace.Project) bool {
	var okGenerators []docspace.DocGenerator
	generators := docspace.GetGenerators()
	if len(project.Config.Generator) == 0 {
		showWarn("please use at least one generator \n")
		return false
	}

	for _, name := range project.Config.Generator {
		if generators[name] == nil {
			showErr("generator \"%s\" is not found,you can choose generator below :\n", name)
			for n := range generators {
				fmt.Printf("    %s\n", n)
			}
			return false
		}
		okGenerators = append(okGenerators, generators[name])
	}
	project.Generators = okGenerators
	return true
}

func checkConfig(config *docspace.Config) error {
	// check if the pkg to scan is exist
	if config.Package == "" {
		return fmt.Errorf("please config a pkg to scan in conf.yaml")
	}

	if config.UseGOModule {
		path := config.Package
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
			if _, err := os.Stat(filepath.Join(gopath, "src", config.Package)); err == nil {
				pkgExist = true
			}
		}
		if !pkgExist {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("error: package \"%s\" is not found in any of:\n", config.Package))
			for _, gopath := range goPaths {
				sb.WriteString(fmt.Sprintln("  ", filepath.Join(gopath, "src", config.Package)))
			}
			return fmt.Errorf("%s", sb.String())
		}
	}
	return nil
}

func scanAPIs(project *docspace.Project) ([]*docspace.API, error) {
	var apis []*docspace.API
	for _, scanner := range project.Scanners {
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

func initGoModule(project *docspace.Project) error {
	data, err := ioutil.ReadFile(filepath.Join(project.Config.Package, "go.mod"))
	if err != nil {
		return err
	}
	project.ModulePkg = docspace.ModulePath(data)
	project.ModulePath = docspace.FindGOModAbsPath(project.Config.Package)
	return nil
}

func makeDoc(ctx *kingpin.ParseContext) error {
	config, err := readConfig()
	if err != nil {
		return showErr("fail to read config file: %v", err)
	}

	if err := checkConfig(config); err != nil {
		return showErr("%v", err)
	}

	project := &docspace.Project{
		Config: config,
	}

	if config.UseGOModule {
		if err := initGoModule(project); err != nil {
			return showErr("%v", err)
		}
	}

	if ok := checkScanner(project); !ok {
		return nil
	}

	if ok := checkGenerator(project); !ok {
		return nil
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

	if tag == "" {
		tag = "all"
	}

	genCtx := &docspace.DocGenContext{
		Tag:    tag,
		APIs:   matchedAPIs,
		Config: *config,
	}

	err = gen(project, genCtx)
	if err != nil {
		return showErr("%v", err)
	}

	fmt.Printf("🍺  done!\n")
	return nil
}

func gen(project *docspace.Project, ctx *docspace.DocGenContext) error {
	var version string
	if makeDocVersion != nil && *makeDocVersion != "" {
		version = *makeDocVersion
	}
	docName := ctx.Tag
	if version != "" {
		docName += "_" + version
	}
	for _, generator := range project.Generators {
		out, err := generator.Gen(ctx)
		if err != nil {
			return err
		}
		fileName := docName + "." + generator.FileExt()
		err = writeFile(generator.Name(), fileName, out)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFile(dir, name string, data []byte) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Printf("📖  write api doc to './docs/%s/%s'\n", dir, name)
	mdPath := filepath.Join(path, "docs", dir)
	if _, err = os.Stat(mdPath); err != nil {
		err = os.Mkdir(mdPath, 0755)
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile(filepath.Join(mdPath, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("witre file ,%v\n", err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("witre file ,%v\n", err)
	}
	return nil
}
