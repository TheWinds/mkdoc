package main

import (
	"go/ast"
	"strings"
)

type structField struct {
	Name       string
	Type       string
	Comment    string
	DocComment string
	JSONTag    string
}

type structInfo struct {
	Name     string
	Fields   []structField
	FieldNum int
}

// 获取结构体信息
func findStructInfo(structName string, f *ast.Package) (*structInfo, error) {

	info := new(structInfo)
	info.Fields = make([]structField, 0)
	// 从语法树获取内容
	ast.Inspect(f, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.TypeSpec:
			structNode := node.(*ast.TypeSpec)
			if structNode.Name.Name == structName {
				info.Name = structNode.Name.Name
				structFields := (structNode.Type).(*ast.StructType).Fields
				for _, field := range structFields.List {
					name := field.Names[0].Name
					comment := ""
					if field.Comment != nil && len(field.Comment.List) != 0 {
						comment = (field.Comment.List[0]).Text
					}
					var tag string
					if field.Tag != nil {
						tag = getJSONTag(field.Tag.Value, name)
					}
					info.FieldNum++
					info.Fields = append(info.Fields, structField{
						Name:       name,
						Comment:    comment,
						DocComment: field.Doc.Text(),
						JSONTag:    tag,
					})

				}
			}

		}
		return true
	})
	if info.Name == "" {
		return nil, nil
	}
	return info, nil
}

func getJSONTag(tags, defaultTag string) string {
	if !strings.Contains(tags, "json") {
		return defaultTag
	}
	return getMidString(tags, "json:\"", "\"")
}

func getMidString(src, s, e string) string {
	sIndex := strings.Index(src, s)
	eIndex := strings.Index(src[sIndex+len(s)+1:], e) + len(s) + sIndex
	return src[sIndex+len(s) : eIndex+1]
}
