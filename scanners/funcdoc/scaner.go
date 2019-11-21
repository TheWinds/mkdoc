package funcdoc

import (
	"docspace"
	"docspace/scanners"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

func init() {
	docspace.RegisterScanner(&Scanner{})
}

type Scanner struct{}

func (c *Scanner) ScanAnnotations(pkg string) ([]docspace.DocAnnotation, error) {
	srcPaths := docspace.GetGOSrcPaths()

	var allDirs []string
	for _, srcPath := range srcPaths {
		rootDir := filepath.Join(srcPath, pkg)
		subDirs := scanners.GetSubDirs(rootDir)
		// filter path
		for _, dir := range subDirs {
			allDirs = append(allDirs, dir)
		}
	}
	fileImports := map[string]map[string]string{}

	annotations := make([]docspace.DocAnnotation, 0)
	for _, dir := range allDirs {
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
						sb := strings.Builder{}
						sb.WriteString(funcNode.Doc.Text())
						sb.WriteString(fmt.Sprintf("@apidoc type http\n"))
						fileName := f.Position(node.Pos()).Filename
						for name, path := range fileImports[fileName] {
							sb.WriteString(fmt.Sprintf("@apidoc pkg_map %s %s\n", name, path))
						}
						annotations = append(annotations, docspace.DocAnnotation(sb.String()))
					}
				}
				return true
			})
		}
	}
	return annotations, nil
}

func (c *Scanner) GetName() string {
	return "funcdoc"
}

func (c *Scanner) SetConfig(map[string]interface{}) {
}

func (c *Scanner) GetHelp() string {
	return "scan doc annotation from go function doc"
}
