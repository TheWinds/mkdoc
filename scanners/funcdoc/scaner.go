package funcdoc

import (
	"docspace"
	"docspace/scanners"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
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

	annotations := make([]docspace.DocAnnotation, 0)
	for _, dir := range allDirs {
		f := token.NewFileSet()
		pkgs, err := parser.ParseDir(f, dir, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		for _, v := range pkgs {
			ast.Inspect(v, func(node ast.Node) bool {
				if funcNode, ok := node.(*ast.FuncDecl); ok {
					if annotation := docspace.GetAnnotationFromComment(funcNode.Doc.Text()); annotation != "" {
						annotation = annotation.AppendMetaData("http", f.Position(funcNode.Doc.Pos()))
						annotations = append(annotations, annotation)
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
