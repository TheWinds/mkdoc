package docspace

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type goStructField struct {
	Name       string
	Type       string
	Comment    string
	DocComment string
	JSONTag    string
}

type goStructInfo struct {
	Name     string
	Fields   []goStructField
	FieldNum int
}

var ErrGoStructNotFound error = errors.New("go struct not found")

// 从语法树获取结构体信息
func findGOStructInfo(structName string, pkg *ast.Package, fileset *token.FileSet) (*goStructInfo, error) {

	info := new(goStructInfo)
	info.Fields = make([]goStructField, 0)
	// 从语法树获取内容
	ast.Inspect(pkg, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.TypeSpec:
			structNode := node.(*ast.TypeSpec)
			if structNode.Name.Name == structName {
				info.Name = structNode.Name.Name
				switch structNode.Type.(type) {
				case *ast.StructType:
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
						baseTyp := baseTypeName(field.Type)
						fileName := fileset.File(node.Pos()).Name()
						for k, v := range pkg.Files[fileName].Imports {
							fmt.Println(k, v.Name,v.Path.Value)
						}
						typ := ""
						if !baseTyp.NotSupport {
							if baseTyp.IsRep {
								typ += "[]"
							}
							if baseTyp.IsRef {
								typ += "*"
							}
							if baseTyp.PkgName != "" {
								typ += baseTyp.PkgName + "." + baseTyp.Name
							} else {
								typ += baseTyp.Name
							}
						}
						info.Fields = append(info.Fields, goStructField{
							Name:       name,
							Comment:    comment,
							DocComment: field.Doc.Text(),
							JSONTag:    tag,
							Type:       typ,
						})

					}
				case *ast.Ident:
					structNode := (structNode.Type).(*ast.Ident)
					info.Name = structNode.Name
				default:
					fmt.Printf("WARNING: only support `type <TypeName> <StructName>` go syntax,plase check %s \n", info.Name)
				}

			}

		}
		return true
	})
	if info.Name == "" {
		return nil, ErrGoStructNotFound
	}
	return info, nil
}

func getJSONTag(tags, defaultTag string) string {
	if !strings.Contains(tags, "json") {
		return defaultTag
	}
	return strings.Replace(getMidString(tags, "json:\"", "\""), ",omitempty", "", -1)
}

func getMidString(src, s, e string) string {
	sIndex := strings.Index(src, s)
	eIndex := strings.Index(src[sIndex+len(s)+1:], e) + len(s) + sIndex
	return src[sIndex+len(s) : eIndex+1]
}

type baseType struct {
	Name       string
	IsRep      bool
	IsRef      bool
	PkgName    string
	NotSupport bool
}

func baseTypeName(x ast.Expr) *baseType {
	switch t := x.(type) {
	case *ast.Ident:
		return &baseType{Name: t.Name}
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			// only possible for qualified type names;
			// assume type is imported
			return &baseType{Name: t.Sel.Name, PkgName: t.X.(*ast.Ident).Name}
		}
	case *ast.ParenExpr:
		return baseTypeName(t.X)
	case *ast.StarExpr:
		bt := baseTypeName(t.X)
		bt.IsRef = true
		return bt
	case *ast.ArrayType:
		bt := baseTypeName(t.Elt)
		bt.IsRep = true
		return bt
	}
	return &baseType{NotSupport: true}
}
