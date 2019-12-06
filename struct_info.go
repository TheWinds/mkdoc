package docspace

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
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

type StructFinder struct{}

type walkCtx struct {
	structName string
	pkg        *ast.Package
	fileset    *token.FileSet
	finder     *StructFinder
	result     *GoStructInfo
	err        error
}

func (s *StructFinder) Find(pkgDir string, structName string) (*GoStructInfo, error) {
	ctx := &walkCtx{
		structName: structName,
		finder:     s,
		result:     new(GoStructInfo),
		fileset:    token.NewFileSet(),
	}
	ctx.result.Fields = make([]*GoStructField, 0)

	pkgs, err := parser.ParseDir(ctx.fileset, pkgDir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		ctx.pkg = pkg
		ast.Inspect(pkg, s.genWalkForStruct(ctx))
	}
	if ctx.err != nil {
		return nil, err
	}
	if ctx.result.Name == "" {
		return nil, errGoStructNotFound
	}
	return ctx.result, nil
}

func (s *StructFinder) genWalkForStruct(ctx *walkCtx) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		switch t := node.(type) {
		case *ast.TypeSpec:
			if t.Name.Name == ctx.structName {
				s.walkTypeSpec(t, ctx)
			}
		}
		return true
	}
}

func (s *StructFinder) walkTypeSpec(spec *ast.TypeSpec, ctx *walkCtx) {
	ctx.result.Name = spec.Name.Name
	switch t := spec.Type.(type) {
	case *ast.StructType:
		fields := t.Fields
		for _, field := range fields.List {
			name := field.Names[0].Name
			var comment string

			if field.Comment != nil && len(field.Comment.List) != 0 {
				comment = (field.Comment.List[0]).Text
			}

			baseTyp := baseType(field.Type)
			imports := GetFileImportsAtNode(spec, ctx.pkg, ctx.fileset)
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
			ctx.result.Fields = append(ctx.result.Fields, structField)

		}
	case *ast.Ident:
		ctx.result.Name = t.Name
	default:
		fmt.Printf("WARNING: only support `type <TypeName> <StructName>` go syntax,plase check %s \n", ctx.result.Name)
	}
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
