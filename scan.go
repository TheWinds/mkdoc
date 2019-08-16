package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
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
				fmt.Println(fileName, file.Name)
				fmt.Println("imports:")
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
					fmt.Println("  ", importName, importPath)
				}
				fileImports[file.Name.Name] = fImportFilesMap
			}
		}

		for _, v := range pkgs {
			ast.Inspect(v, func(node ast.Node) bool {
				if kvExpr, ok := node.(*ast.KeyValueExpr); ok {
					name := readCode(f, kvExpr.Key)
					if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
						value := readCode(f, kvExpr.Value)
						if strings.HasPrefix(value, "&graphql.Field{") {
							fmt.Println("key:", name)
							fmt.Println("file:", f.Position(kvExpr.Pos()).Filename)
							fmt.Println(value)
							fmt.Println()
							count++
						}

					}
				}
				return true
			})
		}

	}
	fmt.Println("Count:", count)
	return nil, nil
}

func parseDocDefFromGraphqlSrc(resolveName string, code string, imports map[string]string) (string, error) {
	graphQLResolveSource := &GraphQLResolveSource{
		ResolveName: resolveName,
		Code:        code,
		Imports:     imports,
	}
	graphQLResolveSource.GetType()
	return "", nil
}

type GraphQLResolveSource struct {
	ResolveName string
	Code        string
	Imports     map[string]string
}

func (g *GraphQLResolveSource) GetType() {

	typeConvFuncNames := []string{
		"NewGraphQLListTypeFromRPCType",
		"NewGraphQLTypeFromRPCType",
		"NewGraphQLArgsFromRPCType"}
	lines := strings.Split(g.Code, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Type") {

		}
		if strings.Contains(line, "Arg") {

		}
		if strings.Contains(line, "@apidoc") {

		}
	}

}

func (g *GraphQLResolveSource) getGoType(codeLine string) string {
	
	if strings.Contains(codeLine, "Type") {

	}
	if strings.Contains(codeLine, "Arg") {

	}
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
