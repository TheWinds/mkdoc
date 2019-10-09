package docspace

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

type GoStructField struct {
	Name       string
	Comment    string
	DocComment string
	JSONTag    string
	GoType     *GoType
}

type GoStructInfo struct {
	Name   string
	Fields []*GoStructField
}

var ErrGoStructNotFound error = errors.New("go struct not found")

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
						var comment, tag string

						if field.Comment != nil && len(field.Comment.List) != 0 {
							comment = (field.Comment.List[0]).Text
						}

						if field.Tag != nil {
							tag = getJSONTag(field.Tag.Value, name)
						}
						baseTyp := baseType(field.Type)
						imports := getNodeFileImports(node, pkg, fileset)
						baseTyp.ImportPkgName = imports[baseTyp.PkgName]

						info.Fields = append(info.Fields, &GoStructField{
							Name:       name,
							Comment:    comment,
							DocComment: field.Doc.Text(),
							JSONTag:    tag,
							GoType:     baseTyp,
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

var importCache map[string]map[string]string

func getNodeFileImports(node ast.Node, pkg *ast.Package, fileset *token.FileSet) map[string]string {
	if importCache == nil {
		importCache = make(map[string]map[string]string)
	}
	fileName := fileset.File(node.Pos()).Name()
	if importCache[fileName] == nil {
		importCache[fileName] = make(map[string]string)
		for _, v := range pkg.Files[fileName].Imports {
			importName := ""
			importPath := strings.Replace(v.Path.Value, "\"", "", -1)
			if v.Name != nil {
				importName = v.Name.Name
			} else {
				importName = filepath.Base(importPath)
			}
			importCache[fileName][importName] = importPath
		}
		importCache[fileName][""] = getPkgPath(fileName)
	}
	return importCache[fileName]
}

func getPkgPath(fileName string) string {
	goPaths := GetGOPaths()
	for _, v := range goPaths {
		v := v + "/src/"
		if strings.HasPrefix(fileName, v) {
			return filepath.Dir(fileName[len(v):])
		}
	}
	return ""
}

func getMidString(src, s, e string) string {
	sIndex := strings.Index(src, s)
	eIndex := strings.Index(src[sIndex+len(s)+1:], e) + len(s) + sIndex
	return src[sIndex+len(s) : eIndex+1]
}

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
	}
	return &GoType{NotSupport: true}
}
