package gofunc

import (
	"github.com/thewinds/mkdoc"
	"go/ast"
	"strings"
)

func init() {
	mkdoc.RegisterDocScanner(new(Scanner))
}

type Scanner struct {
	enableGoMod bool
	filterTag   string
	pkg         string
}

func (s *Scanner) Scan(config mkdoc.DocScanConfig) (*mkdoc.DocScanResult, error) {
	if config.Args[EnableGoModule] == "true" {
		s.enableGoMod = true
	}
	s.filterTag = config.Args["_filter_tag"]
	if len(config.Args["pkg"]) > 0 {
		s.pkg = config.Args["pkg"]
	} else {
		s.pkg = config.Args["path"]
	}
	if err := mkdoc.CheckGoScanPath(s.pkg, s.enableGoMod); err != nil {
		return nil, err
	}
	annotations, err := s.scanAnnotations()
	if err != nil {
		return nil, err
	}
	r := new(mkdoc.DocScanResult)
	for _, v := range annotations {
		api, err := parseSimple(v)
		if err != nil {
			return nil, err
		}
		if len(s.filterTag) > 0 {
			var ok bool
			for _, tag := range api.Tags {
				if tag == strings.TrimSpace(s.filterTag) {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		objects, err := parseInOut(v, api)
		if err != nil {
			return nil, err
		}
		r.APIs = append(r.APIs, api)
		r.Objects = append(r.Objects, objects...)
	}
	return r, nil
}

func (s *Scanner) scanAnnotations() ([]DocAnnotation, error) {
	annotations := make([]DocAnnotation, 0)
	dirs := mkdoc.GetScanDirs(s.pkg, s.enableGoMod, nil)
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
