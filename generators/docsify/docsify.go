package docsify

import (
	"bytes"
	"fmt"
	"github.com/thewinds/mkdoc"
	"github.com/thewinds/mkdoc/generators/objmock"
	"sort"
	"strings"
	"time"
)

type Generator struct {
	tagAPIs map[string][]*mkdoc.API
	tags    []string
	refObj  map[string]*mkdoc.Object
}

func (g *Generator) Gen(ctx *mkdoc.DocGenContext) (output *mkdoc.GeneratedOutput, err error) {
	g.refObj = ctx.RefObj
	g.groupAPIByTag(ctx)
	output = &mkdoc.GeneratedOutput{Files: []*mkdoc.GeneratedFile{
		g.makeIndex(ctx),
		g.makeReadme(ctx),
		g.makeSidebar(ctx),
	}}
	for _, tag := range g.tags {
		md, err := g.makeTagMD(tag)
		if err != nil {
			return nil, err
		}
		output.Files = append(output.Files, md)
	}
	return
}

func (g *Generator) makeSidebar(ctx *mkdoc.DocGenContext) *mkdoc.GeneratedFile {
	buf := bytes.NewBuffer(nil)
	writeLine := func(s string) {
		buf.WriteString(s)
		buf.WriteByte('\n')
	}
	writeLine("- Getting started")
	writeLine("  - [README](/)")
	writeLine("")
	writeLine("- APIs")
	for _, tag := range g.tags {
		writeLine(fmt.Sprintf("  - [%s](%s.md)", tag, tag))
	}
	return &mkdoc.GeneratedFile{Name: "_sidebar.md", Data: buf.Bytes()}
}

const indexTpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>%s</title>
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1" />
  <meta name="description" content="Description">
  <meta name="viewport" content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
  <link rel="stylesheet" href="//unpkg.com/docsify/lib/themes/vue.css">
</head>
<body>
  <nav>
    <a href="#" style="color:#42b983">UpdateAt: %s</a>
  </nav>
  <div id="app"></div>
  <script>
    window.$docsify = {
      name: '',
      repo: ''
    }
  </script>
  <script>
    window.$docsify = {
      loadSidebar: true,
      subMaxLevel: 2,
      search: 'auto',
    }
  </script>
  <script src="//unpkg.com/docsify/lib/docsify.min.js"></script>
  <script src="//cdn.jsdelivr.net/npm/docsify/lib/plugins/search.min.js"></script>
  <script src="//cdn.jsdelivr.net/npm/docsify-copy-code"></script>
  <script src="//cdn.jsdelivr.net/npm/prismjs/components/prism-json.min.js"></script>
</body>
</html>
`

func (g *Generator) makeIndex(ctx *mkdoc.DocGenContext) *mkdoc.GeneratedFile {
	src := fmt.Sprintf(indexTpl, ctx.Config.Name, time.Now().Format("2006-01-02 15:04:05"))
	return &mkdoc.GeneratedFile{Name: "index.html", Data: []byte(src)}
}

func (g *Generator) makeReadme(ctx *mkdoc.DocGenContext) *mkdoc.GeneratedFile {
	buf := bytes.NewBuffer(nil)
	tpl := `# %s
> %s

> show doc by [docsify](https://github.com/docsifyjs/docsify)
- APIBaseURL: %s `
	buf.WriteString(fmt.Sprintf(tpl,
		ctx.Config.Name,
		ctx.Config.Description,
		ctx.Config.APIBaseURL))

	return &mkdoc.GeneratedFile{Name: "README.md", Data: buf.Bytes()}
}

func (g *Generator) makeTagMD(tag string) (*mkdoc.GeneratedFile, error) {
	markdownBuilder := strings.Builder{}
	writef := func(format string, v ...interface{}) {
		markdownBuilder.WriteString(fmt.Sprintf(format, v...))
	}
	writef("# %s\n\n", tag)
	sort.Slice(g.tagAPIs[tag], func(i, j int) bool {
		return g.tagAPIs[tag][i].Name < g.tagAPIs[tag][j].Name
	})
	for _, api := range g.tagAPIs[tag] {
		writef("## %s\n", api.Name)
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
			o, err := objmock.NewJSONMocker().MockPrettyComment(api.InArgument, g.refObj)
			if err != nil {
				return nil, err
			}
			if len(strings.TrimSpace(o)) == 0 {
				o = "{}"
			}
			writef(o)
		}
		writef("\n```\n\n")

		writef("- Response Example\n")
		writef("```")
		switch api.Mime.Out {
		default:
			writef("json\n")
			o, err := objmock.NewJSONMocker().MockPrettyComment(api.OutArgument, g.refObj)
			if err != nil {
				return nil, err
			}
			if len(strings.TrimSpace(o)) == 0 {
				o = "{}"
			}
			writef(o)
		}
		writef("\n```\n")
	}
	return &mkdoc.GeneratedFile{
		Name: tag + ".md",
		Data: []byte(markdownBuilder.String()),
	}, nil
}

func (g *Generator) groupAPIByTag(ctx *mkdoc.DocGenContext) {
	g.tagAPIs = make(map[string][]*mkdoc.API)
	for _, api := range ctx.APIs {
		for _, tag := range api.Tags {
			if g.tagAPIs[tag] == nil {
				g.tags = append(g.tags, tag)
			}
			g.tagAPIs[tag] = append(g.tagAPIs[tag], api)
		}
	}
	sort.Strings(g.tags)
}

func (g *Generator) Name() string {
	return "docsify"
}

func init() {
	mkdoc.RegisterGenerator(&Generator{})
}
