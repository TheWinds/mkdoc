package gqlcorego

import (
	"errors"
	"fmt"
	"github.com/thewinds/mkdoc"
	"go/ast"
	"go/token"
	"strings"
)

var errDefNotMatch = errors.New("not match graphql-go def ast")

type gqlFieldParser struct {
	FileSet *token.FileSet
	AST     []ast.Expr
}

func (g *gqlFieldParser) Parse() (annotation mkdoc.DocAnnotation, err error) {
	if len(g.AST) == 0 {
		return "", errDefNotMatch
	}

	elts, ok := g.assertGraphQLFieldElts(g.AST)
	if !ok {
		return "", errDefNotMatch
	}

	annotationBuilder := strings.Builder{}

	for _, elt := range elts {
		switch elt.(type) {
		case *ast.KeyValueExpr:
			expr := elt.(*ast.KeyValueExpr)
			keyName := expr.Key.(*ast.Ident).Name
			switch keyName {
			case "Type", "Args":
				if callExpr, ok := expr.Value.(*ast.CallExpr); ok {
					typeConvFuncName := callExpr.Fun.(*ast.SelectorExpr).Sel.Name
					switch callExpr.Args[0].(type) {
					case *ast.CompositeLit:
						typeExpr, ok := (callExpr.Args[0].(*ast.CompositeLit).Type).(*ast.SelectorExpr)
						if !ok {
							return "", errDefNotMatch
						}
						packageName := typeExpr.X.(*ast.Ident).Name
						typeName := typeExpr.Sel.Name
						isRepeated := strings.Contains(typeConvFuncName, "List")
						pkgType := fmt.Sprintf("%s.%s", packageName, typeName)
						if isRepeated {
							pkgType = "[]" + pkgType
						}
						if keyName == "Type" {
							annotationBuilder.WriteString(fmt.Sprintf("@out type %s\n", pkgType))
						} else {
							annotationBuilder.WriteString(fmt.Sprintf("@in type %s\n", pkgType))
						}
					case *ast.SelectorExpr:
					}

				}
				// 通过 FieldConfigArgument 定义的参数类型
				if lit, ok := expr.Value.(*ast.CompositeLit); ok {
					if litType, ok := lit.Type.(*ast.SelectorExpr); ok {
						if litType.Sel.Name == "FieldConfigArgument" {
							annotationBuilder.WriteString(fmt.Sprintf("@in fields {\n"))

							for _, e := range lit.Elts {
								if argKV, ok := e.(*ast.KeyValueExpr); ok {
									fieldName := argKV.Key.(*ast.BasicLit).Value
									annotationBuilder.WriteString(fmt.Sprintf("    %s", strings.Replace(fieldName, "\"", "", -1)))
									for _, fieldAttrElt := range argKV.Value.(*ast.UnaryExpr).X.(*ast.CompositeLit).Elts {
										attrName := fieldAttrElt.(*ast.KeyValueExpr).Key.(*ast.Ident).Name
										switch attrName {
										case "Type":
											fieldTypeSel, ok := fieldAttrElt.(*ast.KeyValueExpr).Value.(*ast.SelectorExpr)
											if !ok {
												return "", errDefNotMatch
											}
											fieldType := g.mapToGoType(fieldTypeSel.Sel.Name)
											annotationBuilder.WriteString(fmt.Sprintf(" %s", fieldType))
										case "Description":
											fieldComment := fieldAttrElt.(*ast.KeyValueExpr).Value.(*ast.BasicLit).Value
											annotationBuilder.WriteString(fmt.Sprintf(" %s", strings.Replace(fieldComment, "\"", "", -1)))
										}
									}
									annotationBuilder.WriteString(fmt.Sprintf("\n"))
								}
							}
							annotationBuilder.WriteString(fmt.Sprintf("}\n"))
						}
					}
				}
			default:
			}
		default:
		}
	}

	annotationBuilder.WriteString(g.getAnnotationFromCode())

	annotation = mkdoc.DocAnnotation(annotationBuilder.String())
	annotation = annotation.AppendMetaData("graphql", g.FileSet.Position(g.AST[0].Pos()))
	return
}

func (g *gqlFieldParser) assertBasicLit(i interface{}) (*ast.BasicLit, bool) {
	r, ok := i.(*ast.BasicLit)
	return r, ok
}

func (g *gqlFieldParser) assertGraphQLFieldElts(exprs []ast.Expr) ([]ast.Expr, bool) {
	if unaryExpr, ok := exprs[0].(*ast.UnaryExpr); ok {
		if compositeLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
			return compositeLit.Elts, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

func (g *gqlFieldParser) getAnnotationFromCode() string {
	sb := strings.Builder{}
	code := readCode(g.FileSet, g.AST[0])
	lines := strings.Split(code, "\n")
	for k, line := range lines {
		if strings.Contains(line, "@") {
			lineRemoveComment := strings.Replace(line, "//", "", -1)
			sb.WriteString(strings.TrimSpace(lineRemoveComment))
			if k != len(lines)-1 {
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

func (g *gqlFieldParser) mapToGoType(graphQLType string) string {
	m := map[string]string{
		"Int":    "int64",
		"String": "string",
		"Float":  "float",
	}
	return m[graphQLType]
}
