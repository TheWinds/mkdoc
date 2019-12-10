package funcdoc

import (
	"github.com/thewinds/mkdoc"
	"go/ast"
	"go/parser"
	"go/token"
)

func init() {
	mkdoc.RegisterScanner(&Scanner{})
}

type Scanner struct{}

func (c *Scanner) ScanAnnotations(project mkdoc.Project) ([]mkdoc.DocAnnotation, error) {
	annotations := make([]mkdoc.DocAnnotation, 0)
	dirs := mkdoc.GetScanDirs(project.Config.Package, project.Config.UseGOModule, nil)
	for _, dir := range dirs {
		f := token.NewFileSet()
		pkgs, err := parser.ParseDir(f, dir, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		for _, v := range pkgs {
			ast.Inspect(v, func(node ast.Node) bool {
				if funcNode, ok := node.(*ast.FuncDecl); ok {
					if annotation := mkdoc.GetAnnotationFromComment(funcNode.Doc.Text()); annotation != "" {
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
