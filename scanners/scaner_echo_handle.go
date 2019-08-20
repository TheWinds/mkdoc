package scanners

import (
	"docspace"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type CoregoEchoAPIScanner struct {
}

func (c *CoregoEchoAPIScanner) ScanAnnotations(pkg string) ([]docspace.DocAnnotation, error) {
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
	return nil, nil
}

func (c *CoregoEchoAPIScanner) GetName() string {
	return "echo-corego"
}

func (c *CoregoEchoAPIScanner) SetConfig(map[string]interface{}) {
}

func (c *CoregoEchoAPIScanner) GetHelp() string {
	return ""
}
