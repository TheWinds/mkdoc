package markdown

import (
	"docspace"
	"fmt"
	"sort"
	"strings"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) json(api *docspace.API, obj *docspace.Object) string {
	o, _ := newObjJSONMarshaller(api, obj).Marshal()
	return o
}

func (g *Generator) gql(api *docspace.API) string {
	sb := new(strings.Builder)
	ind := strings.LastIndex(api.Path, ":")
	opName := api.Path[ind+1:]
	args := make([]string, 0)
	argsInner := make([]string, 0)
	for _, field := range api.InArgument.Fields {
		var gqlTyp string
		switch field.BaseType {
		case "string":
			gqlTyp = "String!"
		case "bool":
			gqlTyp = "Boolean!"
		case "int", "int32", "int64", "uint", "uint32", "uint64":
			gqlTyp = "Int!"
		case "float", "float32", "float64":
			gqlTyp = "Float!"
		}
		if field.IsRepeated {
			gqlTyp = "[" + gqlTyp + "]!"
		}
		args = append(args, fmt.Sprintf("$%s: %s", field.JSONTag, gqlTyp))
		argsInner = append(argsInner, fmt.Sprintf("%s: $%s", field.JSONTag, field.JSONTag))
	}
	bodykw := "body"
	if api.OutArgumentLoc != nil && api.OutArgumentLoc.IsRepeated {
		bodykw = "bodys"
	}
	ql := `%s %s(%s) {
		%s(%s) {
		  total
		  %s%s
		  errorCode
		  errorMsg
		  success
		}
	  }`
	gqlBody, _ := newObjGQLMarshaller(api, api.OutArgument).Marshal()
	sb.WriteString(
		fmt.Sprintf(
			ql,
			api.Method,
			opName,
			strings.Join(args, ","),
			opName,
			strings.Join(argsInner, ","),
			bodykw,
			gqlBody,
		))

	return sb.String()
}

func (g *Generator) Gen(ctx *docspace.DocGenContext) (output []byte, err error) {
	markdownBuilder := strings.Builder{}
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
		ctx.Config.Name,
		ctx.Config.Description,
		"`"+ctx.Tag+"`",
		fmt.Sprintf("`%d`", len(ctx.APIs))))

	markdownBuilder.WriteString("\n")

	for _, api := range ctx.APIs {
		markdownBuilder.WriteString(fmt.Sprintf("### %s\n", api.Name))
		if len(strings.TrimSpace(api.Desc)) > 0 {
			markdownBuilder.WriteString(fmt.Sprintf("> %s\n", api.Desc))
		}
		markdownBuilder.WriteString("\n")
		markdownBuilder.WriteString(fmt.Sprintf("- %s %s\n", api.Method, api.Type))
		markdownBuilder.WriteString(fmt.Sprintf("```\n"))
		markdownBuilder.WriteString(fmt.Sprintf("[path] %s\n", api.Path))
		markdownBuilder.WriteString(fmt.Sprintf("```\n"))

		if len(api.Header) > 0 {
			markdownBuilder.WriteString("- Header\n")
			markdownBuilder.WriteString("|名称|说明|\n|---|---|\n")
			keys := make([]string, 0, len(api.Header))
			for k := range api.Header {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				markdownBuilder.WriteString(fmt.Sprintf("|`%s`|%s|\n", key, api.Header[key]))
			}
			markdownBuilder.WriteString("\n")
		}

		if len(api.Query) > 0 {
			markdownBuilder.WriteString("- Query\n")
			markdownBuilder.WriteString("|名称|说明|\n|---|---|\n")
			keys := make([]string, 0, len(api.Query))
			for k := range api.Query {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				markdownBuilder.WriteString(fmt.Sprintf("|`%s`|%s|\n", key, api.Query[key]))
			}
			markdownBuilder.WriteString("\n")
		}

		markdownBuilder.WriteString("- Req Body\n")
		markdownBuilder.WriteString(fmt.Sprintf("```json\n"))
		if api.InArgumentLoc != nil && api.InArgumentLoc.IsRepeated {
			markdownBuilder.WriteString(fmt.Sprintf("[\n"))
		}
		markdownBuilder.WriteString(g.json(api, api.InArgument))
		if api.InArgumentLoc != nil && api.InArgumentLoc.IsRepeated {
			markdownBuilder.WriteString(fmt.Sprintf("]\n"))
		} else {
			markdownBuilder.WriteString(fmt.Sprintf("\n"))
		}
		markdownBuilder.WriteString(fmt.Sprintf("```\n"))
		markdownBuilder.WriteString("- Resp Body\n")

		markdownBuilder.WriteString(fmt.Sprintf("```json\n"))
		if api.OutArgumentLoc != nil && api.OutArgumentLoc.IsRepeated {
			markdownBuilder.WriteString(fmt.Sprintf("["))
		}
		markdownBuilder.WriteString(g.json(api, api.OutArgument))
		if api.OutArgumentLoc != nil && api.OutArgumentLoc.IsRepeated {
			markdownBuilder.WriteString(fmt.Sprintf("]\n"))
		} else {
			markdownBuilder.WriteString(fmt.Sprintf("\n"))
		}
		markdownBuilder.WriteString(fmt.Sprintf("```\n"))

		if api.Type == "graphql" {
			//markdownBuilder.WriteString("<details>\n")
			//markdownBuilder.WriteString("<summary>查看GraphQL</summary>")
			markdownBuilder.WriteString(fmt.Sprintf("```\n"))
			markdownBuilder.WriteString(g.gql(api))
			markdownBuilder.WriteString(fmt.Sprintf("\n```\n"))
			//markdownBuilder.WriteString("</details>\n")
		}
		markdownBuilder.WriteString("\n")
	}

	return []byte(markdownBuilder.String()), nil
}

func (g *Generator) Name() string {
	return "markdown"
}
