package scaners

import (
	"go/ast"
	"go/parser"
	"go/token"
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

