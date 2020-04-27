package gofunc

import (
	"github.com/thewinds/mkdoc"
	"go/ast"
)

func init() {
	mkdoc.RegisterDocScanner(new(Scanner))
}

type Scanner struct {
	enableGoMod bool
}

func (s *Scanner) Scan(config mkdoc.DocScanConfig) (*mkdoc.DocScanResult, error) {
	if config.Args[EnableGoModule] == "true" {
		s.enableGoMod = true
	}
	annotations, err := s.scanAnnotations(&config)
	if err != nil {
		return nil, err
	}
	r := new(mkdoc.DocScanResult)
	for _, v := range annotations {
		api, objects, err := v.ParseToAPI()
		if err != nil {
			return nil, err
		}
		r.APIs = append(r.APIs, api)
		r.Objects = append(r.Objects, objects...)
	}
	return r, nil
}

func (s *Scanner) scanAnnotations(config *mkdoc.DocScanConfig) ([]DocAnnotation, error) {
	annotations := make([]DocAnnotation, 0)
	dirs := mkdoc.GetScanDirs(config.ProjectConfig.Path, s.enableGoMod, nil)
	for _, dir := range dirs {

		pkgs, fileset, err := mkdoc.ParseDir(dir)
		if err != nil {
			panic(err)
		}

		for _, v := range pkgs {
			ast.Inspect(v, func(node ast.Node) bool {
				if funcNode, ok := node.(*ast.FuncDecl); ok {
					if annotation := GetAnnotationFromComment(funcNode.Doc.Text()); annotation != "" {
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

func (s *Scanner) Name() string {
	return "gofunc"
}

func (s *Scanner) Help() string {
	return "scan code from go func declare"
}
