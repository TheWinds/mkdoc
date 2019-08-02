package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func getAPIDocFuncInfo(pkg string) {
	f := token.NewFileSet()
	goPath := os.Getenv("GOPATH")
	pkgs, err := parser.ParseDir(f, filepath.Join(goPath, "src", pkg), nil, parser.ParseComments)
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
					for _, v := range funcNode.Body.List {
						switch v.(type) {
						case *ast.AssignStmt:
							printCode(f, v)
						case *ast.DeferStmt:
							printCode(f, v)
						}
					}
				}
			}
			return true
		})
	}

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