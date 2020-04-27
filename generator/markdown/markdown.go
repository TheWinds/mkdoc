package markdown

import (
	"fmt"
	"github.com/thewinds/mkdoc"
	"github.com/thewinds/mkdoc/generator/objmock"
	"sort"
	"strings"
	"time"
)

type Generator struct {
	refObj map[mkdoc.LangObjectId]*mkdoc.Object
}

func init() {
	mkdoc.RegisterGenerator(&Generator{})
}

func (g *Generator) Gen(ctx *mkdoc.DocGenContext) (output *mkdoc.GeneratedOutput, err error) {
	g.refObj = ctx.RefObj
	markdownBuilder := strings.Builder{}
	writef := func(format string, v ...interface{}) {
		markdownBuilder.WriteString(fmt.Sprintf(format, v...))
	}
	header := `
# %s

> %s 

- BaseURL: *%s*

##  Summary

| 📖 **Tag**     | %s |
| ------------- | ------ |
| 🔮 **API Num** | %s   |

[TOC]

# API List
`
	writef(header,
		ctx.Config.Name,
		ctx.Config.Description,
		ctx.Config.APIBaseURL,
		"`"+ctx.Tag+"`",
		fmt.Sprintf("`%d`", len(ctx.APIs)))

	writef("\n")

	for _, api := range ctx.APIs {
		writef("### %s\n", api.Name)
		if len(strings.TrimSpace(api.Desc)) > 0 {
			writef("> %s\n", api.Desc)
		}
		writef("\n")
		writef("- %s %s\n", api.Method, api.Type)
		writef("```\n")
		writef("[path] %s\n", api.Path)
		writef("```\n")

		if len(api.Header) > 0 {
			writef("- Header\n")
			writef("|名称|说明|\n|---|---|\n")
			keys := make([]string, 0, len(api.Header))
			for k := range api.Header {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				writef("|`%s`|%s|\n", key, api.Header[key])
			}
			writef("\n")
		}

		if len(api.Query) > 0 {
			writef("- Query\n")
			writef("|名称|说明|\n|---|---|\n")
			keys := make([]string, 0, len(api.Query))
			for k := range api.Query {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				writef("|`%s`|%s|\n", key, api.Query[key])
			}
			writef("\n")
		}

		writef("- Request Example\n")
		writef("```")
		switch api.Mime.In {
		default:
			writef("json\n")
			o, err := objmock.NewJSONMocker().SetLanguage(api.Language).MockPrettyComment(api.InArgument, ctx.RefObj)
			if err != nil {
				return nil, err
			}
			writef(o)
		}
		writef("\n```\n")

		writef("- Response Example\n")
		writef("```")
		switch api.Mime.Out {
		default:
			writef("json\n")
			o, err := objmock.NewJSONMocker().SetLanguage(api.Language).MockPrettyComment(api.OutArgument, ctx.RefObj)
			if err != nil {
				return nil, err
			}
			writef(o)
		}
		writef("\n```\n")
	}

	var outName string
	if ctx.Tag == "" {
		outName = fmt.Sprintf("all_doc_%s", time.Now().Format("2006_01_02_150405"))
	} else {
		outName = ctx.Tag
	}

	output = &mkdoc.GeneratedOutput{}
	output.Files = append(output.Files, &mkdoc.GeneratedFile{
		Name: outName + ".md",
		Data: []byte(markdownBuilder.String()),
	})
	return output, nil
}

func (g *Generator) Name() string {
	return "markdown"
}
