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

		if api.Method == "query" || api.Method == "mutation" {
			writef("- Graphql Schema\n")
			writef("```")
			switch api.Mime.In {
			default:
				writef("\n")
				o, err := g.gql(api, ctx.RefObj)
				if err != nil {
					return nil, err
				}
				writef(o)
			}
			writef("\n```\n")
		}

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

func (g *Generator) gql(api *mkdoc.API, refs map[mkdoc.LangObjectId]*mkdoc.Object) (string, error) {
	sb := new(strings.Builder)
	ind := strings.LastIndex(api.Path, ":")
	opName := api.Path[ind+1:]
	args := make([]string, 0)
	argsInner := make([]string, 0)
	for _, field := range api.InArgument.Fields {
		fType := field.Type
		if fType.IsRepeated || fType.Name == "object" {
			//fmt.Printf("gql_zk: SKIP '%s' field %s array or object field is not support\n", api.Name, field.Name)
			continue
		}
		goTagExt := getGoTag(field.Extensions)
		if goTagExt.Tag.GetValue("json") == "-" {
			continue
		}
		var jsonTag string
		if goTagExt.Tag.GetValue("json") == "" {
			jsonTag = field.Name
		} else {
			jsonTag = goTagExt.Tag.GetFirstValue("json", ",")
		}
		var gqlTyp string
		switch field.Type.Name {
		case "string":
			gqlTyp = "String!"
		case "bool":
			gqlTyp = "Boolean!"
		case "int", "int32", "int64", "uint", "uint32", "uint64":
			gqlTyp = "Int!"
		case "float", "float32", "float64":
			gqlTyp = "Float!"
		}
		if field.Type.IsRepeated {
			gqlTyp = "[" + gqlTyp + "]!"
		}
		args = append(args, fmt.Sprintf("$%s: %s", jsonTag, gqlTyp))
		argsInner = append(argsInner, fmt.Sprintf("%s: $%s", jsonTag, jsonTag))
	}
	bodykw := "body"
	if api.OutArgument != nil && api.OutArgument.Type.IsRepeated {
		bodykw = "bodys"
	}
	var tArgs string
	var tArgsInner string
	if len(args) > 0 {
		tArgs = fmt.Sprintf("(%s)", strings.Join(args, ","))
	}
	if len(argsInner) > 0 {
		tArgsInner = fmt.Sprintf("(%s)", strings.Join(argsInner, ","))
	}
	ql := `%s %s%s {
		%s%s {
		  total
		  %s%s
		  errorCode
		  errorMsg
		  success
		}
	  }`
	gqlBody, err := objmock.GqlBodyMocker().SetLanguage(api.Language).MockPretty(api.OutArgument, refs, "		  ", "    ")
	if err != nil {
		return "", err
	}
	sb.WriteString(
		fmt.Sprintf(
			ql,
			api.Method,
			opName,
			tArgs,
			opName,
			tArgsInner,
			bodykw,
			gqlBody,
		))

	return sb.String(), nil
}

func (g *Generator) Name() string {
	return "markdown"
}

func getGoTag(exts []mkdoc.Extension) *mkdoc.ExtensionGoTag {
	for _, ext := range exts {
		if e, ok := ext.(*mkdoc.ExtensionGoTag); ok {
			return e
		}
	}
	return nil
}
