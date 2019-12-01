package docspace

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// GoStructField some useful filed info of go struct field
type GoStructField struct {
	Name       string
	Comment    string
	DocComment string
	JSONTag    string `json:"json_tag"`
	XMLTag     string `json:"xml_tag"`
	DocTag     string `json:"doc_tag"`
	GoType     *GoType
}

// GoStructInfo some useful info of go struct
type GoStructInfo struct {
	Name   string
	Fields []*GoStructField
}

var errGoStructNotFound = errors.New("go struct not found")

// 从语法树获取结构体信息
func findGOStructInfo(structName string, pkg *ast.Package, fileset *token.FileSet) (*GoStructInfo, error) {

	info := new(GoStructInfo)
	info.Fields = make([]*GoStructField, 0)
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
						var comment string

						if field.Comment != nil && len(field.Comment.List) != 0 {
							comment = (field.Comment.List[0]).Text
						}

						baseTyp := baseType(field.Type)
						imports := GetFileImportsAtNode(node, pkg, fileset)
						baseTyp.ImportPkgName = imports[baseTyp.PkgName]
						structField := &GoStructField{
							Name:       name,
							Comment:    comment,
							DocComment: field.Doc.Text(),
							GoType:     baseTyp,
						}
						if field.Tag != nil {
							structField.JSONTag = getTag(field.Tag.Value, "json", name)
							structField.XMLTag = getTag(field.Tag.Value, "xml", name)
							structField.DocTag = getTag(field.Tag.Value, "doc", name)
						}
						info.Fields = append(info.Fields, structField)

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
		return nil, errGoStructNotFound
	}
	return info, nil
}

func getTag(tags, tagName, defaultTag string) string {
	if !strings.Contains(tags, tagName) {
		return defaultTag
	}
	v := getMidString(tags, tagName+":\"", "\"")
	i := strings.Index(v, ",")
	if i == -1 {
		return v
	}
	return v[:i]
}

func getMidString(src, s, e string) string {
	sIndex := strings.Index(src, s)
	eIndex := strings.Index(src[sIndex+len(s)+1:], e) + len(s) + sIndex
	return src[sIndex+len(s) : eIndex+1]
}

// GoType describe a go type from go ast
type GoType struct {
	Name          string
	IsRep         bool
	IsRef         bool
	PkgName       string
	ImportPkgName string
	NotSupport    bool
}

func (t *GoType) String() string {
	var typ, importInfo string
	if !t.NotSupport {
		if t.IsRep {
			typ += "[]"
		}
		if t.IsRef {
			typ += "*"
		}
		if t.PkgName != "" {
			typ += t.PkgName + "." + t.Name
		} else {
			typ += t.Name
		}
	}
	if t.ImportPkgName != "" {
		importInfo += t.PkgName + " => " + t.ImportPkgName
	}
	return fmt.Sprintf("Name: %s\nIsRef: %v\nImport:%s", typ, t.IsRef, importInfo)
}

// Location return the location info of go type
func (t *GoType) Location() *TypeLocation {
	return &TypeLocation{
		PackageName: t.ImportPkgName,
		TypeName:    t.Name,
		IsRepeated:  t.IsRep,
	}
}

func baseType(x ast.Expr) *GoType {
	switch t := x.(type) {
	case *ast.Ident:
		return &GoType{Name: t.Name}
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			return &GoType{Name: t.Sel.Name, PkgName: t.X.(*ast.Ident).Name}
		}
	case *ast.ParenExpr:
		return baseType(t.X)
	case *ast.StarExpr:
		bt := baseType(t.X)
		bt.IsRef = true
		return bt
	case *ast.ArrayType:
		bt := baseType(t.Elt)
		bt.IsRep = true
		return bt
	case *ast.InterfaceType:
		return &GoType{Name: "interface{}"}
	}
	return &GoType{NotSupport: true}
}
