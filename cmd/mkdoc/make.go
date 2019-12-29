// command: mkdoc make
package mkdoc

import (
	"fmt"
	"github.com/thewinds/mkdoc"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"sort"
)

func scanAPIs(project *mkdoc.Project) ([]*mkdoc.API, error) {
	var apis []*mkdoc.API
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
			apis = append(apis, api)
		}
		fmt.Printf("\n")
	}
	return apis, nil
}

func getAllTags(apis []*mkdoc.API) []string {
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

func filterAPIByTag(apis []*mkdoc.API, tag string) []*mkdoc.API {
	var matched []*mkdoc.API

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

func buildAPI(apis []*mkdoc.API) error {
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
	config, err := mkdoc.LoadDefaultConfig()
	if err != nil {
		return showErr("fail to read config file: %v", err)
	}

	project, err := mkdoc.NewProject(config)
	if err != nil {
		return showErr("%v", err)
	}
	mkdoc.SetProject(project)

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

	if err := project.LoadObjects(); err != nil {
		return showErr("%v", err)
	}

	if err := buildAPI(matchedAPIs); err != nil {
		return showErr("%v", err)
	}

	if tag == "" {
		tag = "all"
	}

	genCtx := &mkdoc.DocGenContext{
		Tag:    tag,
		APIs:   matchedAPIs,
		Config: *config,
		RefObj: project.Objects(),
	}

	err = gen(project, genCtx)
	if err != nil {
		return showErr("%v", err)
	}

	fmt.Printf("🍺  done!\n")
	return nil
}

func gen(project *mkdoc.Project, ctx *mkdoc.DocGenContext) error {
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
