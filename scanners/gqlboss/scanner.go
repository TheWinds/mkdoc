package gqlboss

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/thewinds/mkdoc"
)

func init() {
	mkdoc.RegisterScanner(&Scanner{})
}

var regSchemaPath = regexp.MustCompile(`path\s+(.+)`)

type Scanner struct {
	fileSet         *token.FileSet
	currentPkg      *ast.Package
	filedAnnotation map[string]mkdoc.DocAnnotation
	fieldSchema     map[string]opSchema
	currentSchema   string //deep first
	schemaPath      map[string]string
	err             error
}

type opSchema struct {
	OpName     string
	SchemaName string
}

func (s *Scanner) walkNode(node ast.Node) bool {
	if kvExpr, ok := node.(*ast.KeyValueExpr); ok {
		if kvExpr.Key == nil || kvExpr.Value == nil {
			return true
		}
		name := readCode(s.fileSet, kvExpr.Key)
		if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
			value := readCode(s.fileSet, kvExpr.Value)
			if strings.HasSuffix(value, "Field()") {
				funcName := value[:len(value)-2]
				opName := name[1:]
				opName = opName[:len(opName)-1]
				s.fieldSchema[funcName] = opSchema{OpName: opName, SchemaName: s.currentSchema}
			}
		}
		return true
	}

	if funcDecl, ok := node.(*ast.FuncDecl); ok {
		funcName := funcDecl.Name.Name
		if funcDecl.Type.Results == nil {
			return false
		}
		retTypeName := readCode(s.fileSet, funcDecl.Type.Results)
		switch retTypeName {
		case "*graphql.Field":
			stmts := funcDecl.Body.List
			for _, stmt := range stmts {
				if returnStmt, ok := stmt.(*ast.ReturnStmt); ok {
					code := readCode(s.fileSet, returnStmt)
					if !strings.Contains(code, "@doc") {
						continue
					}
					p := &gqlFieldParser{
						FileSet: s.fileSet,
						AST:     returnStmt.Results,
					}
					annotation, err := p.Parse()
					if err != nil {
						if err == errDefNotMatch {
							continue
						}
						s.err = err
						return false
					}
					s.filedAnnotation[funcName] = annotation
					return true
				}
			}
		case "graphql.Schema", "*graphql.Schema":
			if retTypeName == "graphql.Schema" {
				s.currentSchema = funcName
				var comments []*ast.Comment
				if funcDecl.Doc != nil {
					comments = funcDecl.Doc.List
				}
				if len(comments) > 0 {
					matches := regSchemaPath.FindStringSubmatch(comments[0].Text)
					if len(matches) == 2 {
						s.schemaPath[funcName] = matches[1]
					}
				}
			}
		}
	}
	return true
}

func (s *Scanner) ScanAnnotations(project mkdoc.Project) ([]mkdoc.DocAnnotation, error) {
	dirs := mkdoc.GetScanDirs(
		project.Config.Package,
		project.Config.UseGOModule,
		func(dirName string) bool {
			return strings.Contains(dirName, "service/boss/schemas")
		})

	s.fileSet = token.NewFileSet()
	s.filedAnnotation = make(map[string]mkdoc.DocAnnotation)
	s.fieldSchema = make(map[string]opSchema)
	s.schemaPath = make(map[string]string)

	for _, dir := range dirs {
		pkgs, err := parser.ParseDir(s.fileSet, dir, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		for _, v := range pkgs {
			s.currentPkg = v
			ast.Inspect(v, s.walkNode)
			if s.err != nil {
				return nil, s.err
			}
		}
	}
	var annotations []mkdoc.DocAnnotation
	for filedFuncName, annotation := range s.filedAnnotation {
		opSchema := s.fieldSchema[filedFuncName]
		pathdoc := fmt.Sprintf("@path %s:%s\n", s.schemaPath[opSchema.SchemaName], opSchema.OpName)
		annotations = append(annotations, annotation+mkdoc.DocAnnotation(pathdoc))
	}
	return annotations, nil

}

// readCode read source code from ast.Node
func readCode(f *token.FileSet, node ast.Node) string {
	ps := f.Position(node.Pos())
	pe := f.Position(node.End())
	file, _ := ioutil.ReadFile(ps.Filename)
	return string(file[ps.Offset:pe.Offset])
}

func (s *Scanner) GetName() string {
	return "gqlboss"
}

func (s *Scanner) SetConfig(map[string]interface{}) {}

func (s *Scanner) GetHelp() string {
	return "scan doc annotation from graphql api in corego/boss project"
}
