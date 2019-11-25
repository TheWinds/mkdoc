package markdown

import (
	"docspace"
	"fmt"
	"sort"
	"strings"
)

type Generator struct {
	api *docspace.API
	sb  *strings.Builder
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Source(api *docspace.API) docspace.DocGenerator {
	g.api = api
	g.sb = new(strings.Builder)
	return g
}

func (g *Generator) write(s string) {
	g.sb.WriteString(s)
}

func (g *Generator) json(obj *docspace.Object) string {
	o, _ := newObjJSONMarshaller(g.api, obj).Marshal()
	return o
}

func (g *Generator) gql() string {
	sb := new(strings.Builder)
	ind := strings.LastIndex(g.api.Path, ":")
	opName := g.api.Path[ind+1:]
	args := make([]string, 0)
	argsInner := make([]string, 0)
	for _, field := range g.api.InArgument.Fields {
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
	if g.api.OutArgumentLoc != nil && g.api.OutArgumentLoc.IsRepeated {
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
	gqlBody, _ := newObjGQLMarshaller(g.api, g.api.OutArgument).Marshal()
	sb.WriteString(
		fmt.Sprintf(
			ql,
			g.api.Method,
			opName,
			strings.Join(args, ","),
			opName,
			strings.Join(argsInner, ","),
			bodykw,
			gqlBody,
		))

	return sb.String()
}

func (g *Generator) Gen() (output string, err error) {
	g.write(fmt.Sprintf("### %s\n", g.api.Name))
	if len(strings.TrimSpace(g.api.Desc)) > 0 {
		g.write(fmt.Sprintf("> %s\n", g.api.Desc))
	}
	g.write("\n")
	g.write(fmt.Sprintf("- %s %s\n", g.api.Method, g.api.Type))
	g.write(fmt.Sprintf("```\n"))
	g.write(fmt.Sprintf("[path] %s\n", g.api.Path))
	g.write(fmt.Sprintf("```\n"))

	if len(g.api.Header) > 0 {
		g.write("- Header\n")
		g.write("|名称|说明|\n|---|---|\n")
		keys := make([]string, 0, len(g.api.Header))
		for k := range g.api.Header {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			g.write(fmt.Sprintf("|`%s`|%s|\n", key, g.api.Header[key]))
		}
		g.write("\n")
	}

	if len(g.api.Query) > 0 {
		g.write("- Query\n")
		g.write("|名称|说明|\n|---|---|\n")
		keys := make([]string, 0, len(g.api.Query))
		for k := range g.api.Query {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			g.write(fmt.Sprintf("|`%s`|%s|\n", key, g.api.Query[key]))
		}
		g.write("\n")
	}

	g.write("- Req Body\n")
	g.write(fmt.Sprintf("```json\n"))
	if g.api.InArgumentLoc != nil && g.api.InArgumentLoc.IsRepeated {
		g.write(fmt.Sprintf("[\n"))
	}
	g.write(g.json(g.api.InArgument))
	if g.api.InArgumentLoc != nil && g.api.InArgumentLoc.IsRepeated {
		g.write(fmt.Sprintf("]\n"))
	} else {
		g.write(fmt.Sprintf("\n"))
	}
	g.write(fmt.Sprintf("```\n"))
	g.write("- Resp Body\n")

	g.write(fmt.Sprintf("```json\n"))
	if g.api.OutArgumentLoc != nil && g.api.OutArgumentLoc.IsRepeated {
		g.write(fmt.Sprintf("["))
	}
	g.write(g.json(g.api.OutArgument))
	if g.api.OutArgumentLoc != nil && g.api.OutArgumentLoc.IsRepeated {
		g.write(fmt.Sprintf("]\n"))
	} else {
		g.write(fmt.Sprintf("\n"))
	}
	g.write(fmt.Sprintf("```\n"))

	if g.api.Type == "graphql" {
		//g.write("<details>\n")
		//g.write("<summary>查看GraphQL</summary>")
		g.write(fmt.Sprintf("```\n"))
		g.write(g.gql())
		g.write(fmt.Sprintf("\n```\n"))
		//g.write("</details>\n")
	}
	return g.sb.String(), nil
}
