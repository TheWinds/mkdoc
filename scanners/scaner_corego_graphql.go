package scanners

import (
	"docspace"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// 文件 => 路由Path映射
var fileRouterPathMap map[string]string

func init() {
	fileRouterPathMap = map[string]string{
		"corego/service/boss/schemas/adminSchema.go":         "/adminManage",
		"corego/service/boss/schemas/businessSchema.go":      "/channelManage",
		"corego/service/boss/schemas/competionSchema.go":     "/competitionManage",
		"corego/service/boss/schemas/customerSchema.go":      "/customerManage",
		"corego/service/boss/schemas/indexSchema.go":         "/indexManage",
		"corego/service/boss/schemas/manageSchema.go":        "/manage",
		"corego/service/boss/schemas/operationSchema.go":     "/operationManage",
		"corego/service/boss/schemas/payOrderSchema.go":      "/orderManage",
		"corego/service/boss/schemas/zhike/courseSchema.go":  "/zhike/courseManage",
		"corego/service/boss/schemas/zhike/teacherSchame.go": "/zhike/teacherManage",
		"corego/service/boss/schemas/zhike/wordSchema.go":    "/zhike/wordManage",
	}
}

type CoregoGraphQLAPIScanner struct {
}

func (c *CoregoGraphQLAPIScanner) GetName() string {
	return "gql-corego"
}

func (c *CoregoGraphQLAPIScanner) ScanAnnotations(pkg string) ([]docspace.DocAnnotation, error) {
	goPath := os.Getenv("GOPATH")
	//TODO(thewinds):scan all go path
	if strings.Contains(goPath, ":") {
		goPath = strings.TrimSpace(strings.Split(goPath, ":")[0])
	}
	goSrcPath := filepath.Join(goPath, "src") + "/"
	rootDir := filepath.Join(goSrcPath, pkg)
	subDirs := getSubDirs(rootDir)
	fileImports := map[string]map[string]string{}

	annotations := make([]docspace.DocAnnotation, 0)

	for _, dir := range subDirs {
		f := token.NewFileSet()
		pkgs, err := parser.ParseDir(f, dir, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		// 获取包引用关系
		for _, v := range pkgs {
			for fileName, file := range v.Files {
				fImportFilesMap := map[string]string{}
				for _, imp := range file.Imports {
					importName := ""
					importPath := strings.Replace(imp.Path.Value, "\"", "", -1)
					if imp.Name != nil {
						importName = imp.Name.Name
					} else {
						importName = filepath.Base(importPath)
					}
					fImportFilesMap[importName] = importPath
				}
				fileImports[fileName] = fImportFilesMap
			}
		}

		// 从AST提取语法树
		for _, v := range pkgs {
			ast.Inspect(v, func(node ast.Node) bool {
				if kvExpr, ok := node.(*ast.KeyValueExpr); ok {
					name := readCode(f, kvExpr.Key)
					if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
						value := readCode(f, kvExpr.Value)
						if strings.HasPrefix(value, "&graphql.Field{") {
							fileName := f.Position(kvExpr.Pos()).Filename
							if !strings.Contains(value, "@apidoc") {
								return true
							}

							s := &GraphQLResolveSource{
								NodeAST:   kvExpr,
								Code:      value,
								Imports:   fileImports[fileName],
								FileSet:   f,
								GOSrcPath: goSrcPath,
							}
							annotation, err := s.GetDocAnnotation()
							if err != nil {
								return true
							}
							annotations = append(annotations, annotation)
						}
					}
					//return false
				}
				return true
			})
		}

	}

	return annotations, nil
}

func (c *CoregoGraphQLAPIScanner) SetConfig(map[string]interface{}) {
}

func (c *CoregoGraphQLAPIScanner) GetHelp() string {
	return ""
}

type GraphQLResolveSource struct {
	FileSet   *token.FileSet
	NodeAST   *ast.KeyValueExpr
	Code      string
	Imports   map[string]string
	GOSrcPath string
}

func (g *GraphQLResolveSource) GetDocAnnotation() (annotation docspace.DocAnnotation, err error) {
	nodeAPIName, ok := g.assertBasicLit(g.NodeAST.Key)
	if !ok {
		err = fmt.Errorf("not support:key is not string")
		return
	}
	absFileName := g.FileSet.Position(g.NodeAST.Pos()).Filename
	relativeFileName := strings.Replace(absFileName, g.GOSrcPath, "", -1)
	//fmt.Println(relativeFileName)
	annotationBuilder := strings.Builder{}
	annotationBuilder.WriteString(fmt.Sprintf("@apidoc name %s\n", strings.Replace(nodeAPIName.Value, "\"", "", -1)))

	annotationBuilder.WriteString(fmt.Sprintf("@apidoc type graphql\n"))

	path := fileRouterPathMap[relativeFileName]
	if path != "" {
		annotationBuilder.WriteString(fmt.Sprintf("@apidoc path %s\n", path))
	}

	fieldElts, ok := g.assertGraphQLFieldElts(g.NodeAST.Value)
	if !ok {
		err = fmt.Errorf("not support:graphql api must define as a &graphql.Field")
		return
	}

	for _, elt := range fieldElts {
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
						typeExpr := callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr)
						packageName := typeExpr.X.(*ast.Ident).Name
						typeName := typeExpr.Sel.Name
						isRepeated := strings.Contains(typeConvFuncName, "List")
						typeLoc := &docspace.TypeLocation{
							PackageName: g.Imports[packageName],
							TypeName:    typeName,
							IsRepeated:  isRepeated,
						}
						if keyName == "Type" {
							annotationBuilder.WriteString(fmt.Sprintf("@apidoc out gotype %s\n", typeLoc.String()))
						} else {
							annotationBuilder.WriteString(fmt.Sprintf("@apidoc in gotype %s\n", typeLoc.String()))
						}
					case *ast.SelectorExpr:
					}

				}
				// 通过 FieldConfigArgument 定义的参数类型
				if lit, ok := expr.Value.(*ast.CompositeLit); ok {
					if litType, ok := lit.Type.(*ast.SelectorExpr); ok {
						if litType.Sel.Name == "FieldConfigArgument" {
							annotationBuilder.WriteString(fmt.Sprintf("@apidoc in fields {\n"))

							for _, e := range lit.Elts {
								if argKV, ok := e.(*ast.KeyValueExpr); ok {
									fieldName := argKV.Key.(*ast.BasicLit).Value
									annotationBuilder.WriteString(fmt.Sprintf("    %s", strings.Replace(fieldName, "\"", "", -1)))
									for _, fieldAttrElt := range argKV.Value.(*ast.UnaryExpr).X.(*ast.CompositeLit).Elts {
										attrName := fieldAttrElt.(*ast.KeyValueExpr).Key.(*ast.Ident).Name
										switch attrName {
										case "Type":
											fieldType := g.mapToGoType(fieldAttrElt.(*ast.KeyValueExpr).Value.(*ast.SelectorExpr).Sel.Name)
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

	annotation = docspace.DocAnnotation(annotationBuilder.String())
	return
}

func (g *GraphQLResolveSource) assertBasicLit(i interface{}) (*ast.BasicLit, bool) {
	r, ok := i.(*ast.BasicLit)
	return r, ok
}

func (g *GraphQLResolveSource) assertGraphQLFieldElts(i interface{}) ([]ast.Expr, bool) {
	if unaryExpr, ok := i.(*ast.UnaryExpr); ok {
		if compositeLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
			return compositeLit.Elts, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

func (g *GraphQLResolveSource) getAnnotationFromCode() string {
	sb := strings.Builder{}
	lines := strings.Split(g.Code, "\n")
	for k, line := range lines {
		if strings.Contains(line, "@apidoc") {
			lineRemoveComment := strings.Replace(line, "//", "", -1)
			sb.WriteString(strings.TrimSpace(lineRemoveComment))
			if k != len(lines)-1 {
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

func (g *GraphQLResolveSource) mapToGoType(graphQLType string) string {
	m := map[string]string{
		"Int":    "int64",
		"String": "string",
		"Float":  "float",
	}
	return m[graphQLType]
}
