// command: mkdoc make
package main

import (
	"fmt"
	"github.com/thewinds/mkdoc"
	"github.com/thewinds/mkdoc/schema"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"sort"
)

func scanSchemas(project *mkdoc.Project, filterTag string) ([]*schema.Schema, error) {
	var schemas []*schema.Schema
	for _, scanner := range project.Scanners {
		fmt.Printf("🔎  scan doc annotations (use %s)\n", scanner.Name())
		args := project.Config.GetScannerArgs(scanner.Name())
		args["_filter_tag"] = filterTag
		sr, err := scanner.Scan(mkdoc.DocScanConfig{
			ProjectConfig: *project.Config,
			Args:          args,
		})
		if err != nil {
			return nil, fmt.Errorf("scan docs %v\n", err)
		}
		schemas = append(schemas, &schema.Schema{APIs: sr.APIs, Objects: sr.Objects})
	}
	return schemas, nil
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

func makeDoc(ctx *kingpin.ParseContext) error {
	config, err := mkdoc.LoadDefaultConfig()
	if err != nil {
		return showErr("fail to read config file: %v", err)
	}

	project, err := mkdoc.NewProject(config)
	if err != nil {
		return showErr("%v", err)
	}

	tag := *makeDocTag

	schemas, err := scanSchemas(project, tag)
	if err != nil {
		return showErr("%v", err)
	}

	var (
		apiDefs []*schema.API
		apis    []*mkdoc.API
	)
	for _, schemaDef := range schemas {
		err := project.LoadObjects(schemaDef)
		if err != nil {
			return showErr("%v", err)
		}
		for _, api := range schemaDef.APIs {
			apiDefs = append(apiDefs, api)
		}
	}

	if len(apiDefs) == 0 {
		fmt.Printf("👽  no api is matched,all tags:\n")
		// TODO: show all tags
		//for _, t := range getAllTags(schemas) {
		//	fmt.Printf("    %s\n", t)
		//}
		return nil
	}

	if tag != "" {
		fmt.Printf("👽  tag '%s' match %d api\n", tag, len(apiDefs))
	} else {
		fmt.Printf("👽  %d api is matched \n", len(apiDefs))
	}

	for n, def := range apiDefs {
		fmt.Printf("\r🔥 parse & build api '%s' [%d/%d]          ", def.Name, n+1, len(apiDefs))
		a, err := project.ParseSchemaAPI(def)
		if err != nil {
			return showErr("parse api schema %s\n%v\n------\nAt:\n%s:%d\nSource:\n%s\n------\n", def.Name, err, def.SourceFileName, def.SourceLineNum, def.Source)
		}

		if a.InArgument != nil {
			a.InType = a.InArgument.ID
		}
		if a.OutArgument != nil {
			a.OutType = a.OutArgument.ID
		}
		apis = append(apis, a)
	}
	fmt.Println()

	genCtx := &mkdoc.DocGenContext{
		Tag:    tag,
		APIs:   apis,
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
		ctxcp := *ctx
		ctxcp.Args = ctxcp.Config.GetGeneratorArgs(generator.Name())
		out, err := generator.Gen(&ctxcp)
		if err != nil {
			return err
		}
		for _, file := range out.Files {
			err = writeFile(generator.Name(), file.Name, file.Data)
			if err != nil {
				return err
			}
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
	fileName := filepath.Join(path, "docs", dir, name)
	fileDir := filepath.Dir(fileName)
	if _, err = os.Stat(fileDir); err != nil {
		err = os.MkdirAll(fileDir, 0755)
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
