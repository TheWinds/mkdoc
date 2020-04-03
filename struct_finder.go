package mkdoc

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
	Tag        string
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
	}
	ctx.result.Fields = make([]*GoStructField, 0)

	pkgs, fileset, err := ParseDir(pkgDir)
	if err != nil {
		return nil, err
	}

	if ctx.fileset == nil {
		ctx.fileset = fileset
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
			baseTyp.IsBuiltin = isBuiltinType(baseTyp.TypeName)
			structField := &GoStructField{
				Name:       name,
				Comment:    comment,
				DocComment: field.Doc.Text(),
				GoType:     baseTyp,
			}
			structField.DocComment = strings.TrimRight(structField.DocComment, "\n")
			if field.Tag != nil {
				structField.Tag = field.Tag.Value
			}
			ctx.result.Fields = append(ctx.result.Fields, structField)

		}
	case *ast.Ident:
		ctx.result.Name = t.Name
	default:
		fmt.Printf("WARNING: only support `type <TypeName> <StructName>` go syntax,plase check %s \n", ctx.result.Name)
	}
}

// GoType describe a go type from go ast
type GoType struct {
	TypeName      string
	IsArray       bool
	IsRef         bool
	ArrayDepth    int
	PkgName       string
	ImportPkgName string
	NotSupport    bool
	IsBuiltin     bool
}

// Location return the location info of go type
//func (t *GoType) Location() *TypeLocation {
//return &TypeLocation{
//	PackageName: t.ImportPkgName,
//	TypeName:    t.Name,
//	IsRepeated:  t.IsRep,
//}
//}

func baseType(x ast.Expr) *GoType {
	switch t := x.(type) {
	case *ast.Ident:
		return &GoType{TypeName: t.Name}
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			return &GoType{TypeName: t.Sel.Name, PkgName: t.X.(*ast.Ident).Name, IsRef: true}
		}
	case *ast.ParenExpr:
		return baseType(t.X)
	case *ast.StarExpr:
		bt := baseType(t.X)
		return bt
	case *ast.ArrayType:
		bt := baseType(t.Elt)
		bt.IsArray = true
		bt.ArrayDepth++
		return bt
	case *ast.InterfaceType:
		return &GoType{TypeName: "interface{}"}
	}
	return &GoType{NotSupport: true}
}
