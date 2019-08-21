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

type CoregoEchoAPIScanner struct {
}

func (c *CoregoEchoAPIScanner) ScanAnnotations(pkg string) ([]docspace.DocAnnotation, error) {
	goPath := os.Getenv("GOPATH")
	rootDir := filepath.Join(goPath, "src", pkg)
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

		for _, v := range pkgs {
			ast.Inspect(v, func(node ast.Node) bool {
				if funcNode, ok := node.(*ast.FuncDecl); ok {
					if strings.Contains(funcNode.Doc.Text(), "@apidoc") {
						annotationStr := fmt.Sprintf("@apidoc type echo-http\n")
						annotationStr += funcNode.Doc.Text()
						//TODO(thewinds): add import to annotation syntax
						fileName := f.Position(node.Pos()).Filename
						for name, path := range fileImports[fileName] {
							annotationStr = strings.Replace(annotationStr, name+".", path+".", -1)
						}
						annotations = append(annotations, docspace.DocAnnotation(annotationStr))
					}
				}
				return true
			})
		}
	}
	return annotations, nil
}

func (c *CoregoEchoAPIScanner) GetName() string {
	return "echo-corego"
}

func (c *CoregoEchoAPIScanner) SetConfig(map[string]interface{}) {
}

func (c *CoregoEchoAPIScanner) GetHelp() string {
	return ""
}
