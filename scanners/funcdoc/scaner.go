package funcdoc

import (
	"docspace"
	"go/ast"
	"go/parser"
	"go/token"
)

func init() {
	docspace.RegisterScanner(&Scanner{})
}

type Scanner struct{}

func (c *Scanner) ScanAnnotations(project docspace.Project) ([]docspace.DocAnnotation, error) {
	annotations := make([]docspace.DocAnnotation, 0)
	dirs := docspace.GetScanDirs(project.Config.Package, project.Config.UseGOModule, nil)
	for _, dir := range dirs {
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
