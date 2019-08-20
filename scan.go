package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getAPIDocFuncInfo(pkg string) {

	goPath := os.Getenv("GOPATH")
	rootDir := filepath.Join(goPath, "src", pkg)
	subDirs := getSubDirs(rootDir)
	for _, dir := range subDirs {
		println(dir)
		f := token.NewFileSet()
		pkgs, err := parser.ParseDir(f, dir, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		for _, v := range pkgs {
			ast.Inspect(v, func(node ast.Node) bool {
				if funcNode, ok := node.(*ast.FuncDecl); ok {
					if strings.Contains(funcNode.Doc.Text(), "@apidoc") {
						println(funcNode.Name.Name, ":")
						println(funcNode.Doc.Text())
						println("body :")
						printCode(f, funcNode.Body)
						//for _, v := range funcNode.Body.List {
						//	switch v.(type) {
						//	case *ast.AssignStmt:
						//		printCode(f, v)
						//	case *ast.DeferStmt:
						//		printCode(f, v)
						//	}
						//}
					}
				}
				return true
			})
		}
	}

}

func scanGraphQLAPIDocInfo(pkg string) ([]*API, error) {

	goPath := os.Getenv("GOPATH")
	rootDir := filepath.Join(goPath, "src", pkg)
	subDirs := getSubDirs(rootDir)
	count := 0
	fileImports := map[string]map[string]string{}

	for _, dir := range subDirs {
		f := token.NewFileSet()
		pkgs, err := parser.ParseDir(f, dir, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		// get imports
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
							count++
							//ast.Print(f,kvExpr)

							s := &GraphQLResolveSource{
								NodeAST: kvExpr,
								Code:    value,
								Imports: fileImports[fileName],
							}
							s.GetAPI()
						}
					}
					//return false
				}
				return true
			})
		}

	}
	fmt.Println("Count:", count)
	return nil, nil
}

type GraphQLResolveSource struct {
	NodeAST *ast.KeyValueExpr
	Code    string
	Imports map[string]string
}

func assertBasicLit(i interface{}) (*ast.BasicLit, bool) {
	r, ok := i.(*ast.BasicLit)
	return r, ok
}

func assertGraphQLFieldElts(i interface{}) ([]ast.Expr, bool) {
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
func (g *GraphQLResolveSource) GetAPI() (api *API, err error) {
	nodeAPIName, ok := assertBasicLit(g.NodeAST.Key)
	if !ok {
		err = fmt.Errorf("not support:key is not string")
		return
	}
	//fmt.Println("API name:", nodeAPIName.Value)
	api = new(API)
	api.Name = nodeAPIName.Value
	api.Type = "graphql"
	api.Method = "query"
	fieldElts, ok := assertGraphQLFieldElts(g.NodeAST.Value)
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
				//fmt.Println(keyName, ": ")
				//  通过 从GoType 定义的参数类型
				if callExpr, ok := expr.Value.(*ast.CallExpr); ok {
					typeConvFuncName := callExpr.Fun.(*ast.SelectorExpr).Sel.Name
					switch callExpr.Args[0].(type) {
					case *ast.CompositeLit:
						typeExpr := callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr)
						packageName := typeExpr.X.(*ast.Ident).Name
						typeName := typeExpr.Sel.Name
						//fmt.Println("- Fun:", typeConvFuncName)
						//fmt.Println("- PackageName:", packageName)
						//fmt.Println("- TypeName:", typeName)
						isRepeated := strings.Contains(typeConvFuncName, "List")
						typeLoc := &TypeLocation{
							PackageName: g.Imports[packageName],
							TypeName:    typeName,
							IsRepeated:  isRepeated,
						}
						//fmt.Println("- GoType:", typeLoc.String())
						if keyName == "Type" {
							api.outArgumentLoc = typeLoc
						} else {
							api.inArgumentLoc = typeLoc
						}
					case *ast.SelectorExpr:
						//fmt.Println("- TypeName:", "NoName")
						//typeExpr := callExpr.Args[0].(*ast.SelectorExpr)
						//fmt.Println(typeExpr.X.(*ast.Ident).Name)

					}

				}
				// 通过 FieldConfigArgument 定义的参数类型
				if lit, ok := expr.Value.(*ast.CompositeLit); ok {
					if litType, ok := lit.Type.(*ast.SelectorExpr); ok {
						if litType.Sel.Name == "FieldConfigArgument" {
							inObj := &Object{
								ID: fmt.Sprintf("intype.graphql.%s", api.Name),
							}
							fields := make([]*ObjectField, 0)
							for _, e := range lit.Elts {
								if argKV, ok := e.(*ast.KeyValueExpr); ok {
									fieldName := argKV.Key.(*ast.BasicLit).Value
									//fmt.Println("field name:", fieldName)
									field := new(ObjectField)
									field.Name = fieldName
									field.JSONTag = fieldName
									for _, fieldAttrElt := range argKV.Value.(*ast.UnaryExpr).X.(*ast.CompositeLit).Elts {
										attrName := fieldAttrElt.(*ast.KeyValueExpr).Key.(*ast.Ident).Name
										switch attrName {
										case "Type":
											field.Type = g.mapToGoType(fieldAttrElt.(*ast.KeyValueExpr).Value.(*ast.SelectorExpr).Sel.Name)
											//fmt.Println("field type:", field.Type)
										case "Description":
											field.Comment = fieldAttrElt.(*ast.KeyValueExpr).Value.(*ast.BasicLit).Value
											//fmt.Println("field desc:", field.Comment)
										}
									}
									fields = append(fields, field)
								}
							}
							inObj.Fields = fields
							api.InArgument = inObj
						}
					}
				}
			default:
			}
		default:
		}
	}
	err = api.Gen()
	if err != nil {
		return
	}
	api.PrintMarkdown()

	return
	//g.NodeAST.
	typeConvFuncNames := []string{
		"NewGraphQLListTypeFromRPCType",
		"NewGraphQLTypeFromRPCType",
		"NewGraphQLArgsFromRPCType"}
	typeConvFuncNames[0] = ""
	lines := strings.Split(g.Code, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Type") {

		}
		if strings.Contains(line, "Arg") {

		}
		if strings.Contains(line, "@apidoc") {

		}
	}
	return
}

func (g *GraphQLResolveSource) mapToGoType(graphQLType string) string {
	m := map[string]string{
		"Int":    "int64",
		"String": "string",
	}
	return m[graphQLType]
}

func (g *GraphQLResolveSource) getGoType(codeLine string) string {

	if strings.Contains(codeLine, "Type") {

	}
	if strings.Contains(codeLine, "Arg") {

	}
	panic("")
}

func (g *GraphQLResolveSource) getFuncArg(s, funcName string) string {
	s = strings.Replace(s, " ", "", -1)
	return getMidString(s, funcName+"(", ")")
}

func getSubDirs(root string) []string {
	subDirs := make([]string, 0)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			subDirs = append(subDirs, path)
		}
		return nil
	})
	return subDirs
}

func printCode(f *token.FileSet, node ast.Node) {
	println(readCode(f, node))
}

func readCode(f *token.FileSet, node ast.Node) string {
	ps := f.Position(node.Pos())
	pe := f.Position(node.End())
	file, _ := ioutil.ReadFile(ps.Filename)
	return string(file[ps.Offset:pe.Offset])
}

type DocAnnotation string

var commandRe map[string]*regexp.Regexp

func init() {
	commandRe = map[string]*regexp.Regexp{
		"name":             regexp.MustCompile(`(@apidoc\s+name\s+)([^\s]+)`),
		"desc":             regexp.MustCompile(`(@apidoc\s+desc\s+)([^\s]+)`),
		"name_desc":        regexp.MustCompile(`(@apidoc\s+name\s+)([^\s]+)(\s+desc\s+)([^\s]+)`),
		"type":             regexp.MustCompile(`(@apidoc\s+type\s+)([^\s]+)`),
		"path":             regexp.MustCompile(`(@apidoc\s+path\s+)([^\s]+)`),
		"method":           regexp.MustCompile(`(@apidoc\s+method\s+)([^\s]+)`),
		"path_method":      regexp.MustCompile(`(@apidoc\s+path\s+)([^\s]+)(\s+method\s+)([^\s]+)`),
		"tag":              regexp.MustCompile(`(@apidoc\s+tag\s+)([^\s]+)`),
		"in_gotype":        regexp.MustCompile(`(@apidoc\s+in\s+gotype\s+)([^\s]+)`),
		"out_gotype":       regexp.MustCompile(`(@apidoc\s+out\s+gotype\s+)([^\s]+)`),
		"in_fileds_block":  regexp.MustCompile(`(@apidoc\s+in\s+fields\s+(\[\])?{\s+)((.|\s)+)}`),
		"out_fileds_block": regexp.MustCompile(`(@apidoc\s+out\s+fields\s+(\[\])?{\s+)((.|\s)+)}`),
		"filed":            regexp.MustCompile(`(\w+)\s+(\w+)\s*(.+)*`),
	}
}

func (annotation DocAnnotation) ParseToAPI() (*API, error) {
	api := new(API)
	for command, re := range commandRe {
		matchGroups := re.FindStringSubmatch(string(annotation))
		if len(matchGroups) > 0 {
			switch command {
			case "name":
				api.Name = matchGroups[2]
			case "desc":
				api.Desc = matchGroups[2]
			case "name_desc":
				api.Name = matchGroups[2]
				api.Desc = matchGroups[4]
			case "type":
				api.Type = matchGroups[2]
			case "path":
				api.Path = matchGroups[2]
			case "method":
				api.Method = matchGroups[2]
			case "path_method":
				api.Method = matchGroups[2]
				api.Path = matchGroups[4]
			case "tag":
				tagsStr := matchGroups[2]
				api.Tags = make([]string, 0)
				if strings.Contains(tagsStr, ",") {
					for _, tag := range strings.Split(tagsStr, ",") {
						if tag != "" {
							api.Tags = append(api.Tags, strings.TrimSpace(tag))
						}
					}
				} else {
					api.Tags = append(api.Tags, strings.TrimSpace(tagsStr))
				}
			case "in_gotype":
				api.inArgumentLoc = newTypeLocation(matchGroups[2])
			case "out_gotype":
				api.outArgumentLoc = newTypeLocation(matchGroups[2])
			case "in_fileds_block":
				// TODO: isRepeated := matchGroups[2] != ""
				fieldStmts := matchGroups[3]
				api.InArgument = &Object{
					ID:     fmt.Sprintf("#obj_%d", rand.Int63()),
					Fields: parseToObjectFields(fieldStmts),
				}
			case "out_fileds_block":
				fieldStmts := matchGroups[3]
				api.OutArgument = &Object{
					ID:     fmt.Sprintf("#obj_%d", rand.Int63()),
					Fields: parseToObjectFields(fieldStmts),
				}
			}
		}
	}
	return nil, nil
}

func parseToObjectFields(fieldStmts string) []*ObjectField {
	fields := make([]*ObjectField, 0)
	for _, stmt := range strings.Split(fieldStmts, "\n") {
		matchGroups := commandRe["field"].FindStringSubmatch(stmt)
		if len(matchGroups) > 0 {
			if !isBuiltinType(matchGroups[2]) {
				fmt.Printf("type [%s] is not support,skip\n", matchGroups[2])
				continue
			}
			fields = append(fields, &ObjectField{
				Name:       matchGroups[1],
				JSONTag:    matchGroups[1],
				Comment:    strings.TrimSpace(matchGroups[3]),
				Type:       matchGroups[2],
				IsRepeated: false,
				IsRef:      false,
			})
		}
	}
	return fields
}
