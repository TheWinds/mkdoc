package funcdoc

import (
	"github.com/thewinds/mkdoc"
	"github.com/thewinds/mkdoc/objectloader/goloader"
	"github.com/thewinds/mkdoc/scanner/gofunc"
	"go/ast"
)

func init() {
	mkdoc.RegisterScanner(&Scanner{})
}

type Scanner struct{}

func (c *Scanner) ScanAnnotations(project mkdoc.Project) ([]gofunc.DocAnnotation, error) {
	annotations := make([]gofunc.DocAnnotation, 0)
	dirs := goloader.GetScanDirs(project.Config.Package, project.Config.UseGOModule, nil)
	for _, dir := range dirs {

		pkgs, fileset, err := goloader.ParseDir(dir)
		if err != nil {
			panic(err)
		}

		for _, v := range pkgs {
			ast.Inspect(v, func(node ast.Node) bool {
				if funcNode, ok := node.(*ast.FuncDecl); ok {
					if annotation := gofunc.GetAnnotationFromComment(funcNode.Doc.Text()); annotation != "" {
						annotation = annotation.AppendMetaData("http", fileset.Position(funcNode.Doc.Pos()))
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
