package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func getAPIDocFuncInfo(pkg string) {
	f := token.NewFileSet()
	goPath := os.Getenv("GOPATH")
	pkgs, err := parser.ParseDir(f, filepath.Join(goPath, "src", pkg), nil, parser.ParseComments)
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
					for _, v := range funcNode.Body.List {
						switch v.(type) {
						case *ast.AssignStmt:
							printCode(f, v)
						case *ast.DeferStmt:
							printCode(f, v)
						}
					}
				}
			}
			return true
		})
	}

}

func printCode(f *token.FileSet, node ast.Node) {
	println(readCode(f, node))
}

func readCode(f *token.FileSet, node ast.Node) string {
	ps := f.Position(node.Pos())
	pe := f.Position(node.End())
	file, _ := ioutil.ReadFile(ps.Filename)
	return string(file[ps.Offset:pe.Offset])
}



//checkGRPCConnClose("corego/service/xyt/api/zhike/h5")
//checkGRPCConnClose("corego/service/xyt/api/zhike/student")
//checkGRPCConnClose("corego/service/xyt/api/zhike/teacher")
//checkGRPCConnClose("corego/service/zhike-teacher/service")
//checkGRPCConnClose("corego/service/zhike-student/service")
//checkGRPCConnClose("corego/service/boss/schemas")
func checkGRPCConnClose(pkg string) {
	f := token.NewFileSet()
	goPath := os.Getenv("GOPATH")
	pkgs, err := parser.ParseDir(f, filepath.Join(goPath, "src", pkg), nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	for _, v := range pkgs {
		println(v.Name)
		ast.Inspect(v, func(node ast.Node) bool {
			if funcNode, ok := node.(*ast.FuncDecl); ok {
				if strings.HasSuffix(funcNode.Name.Name, "Init") {
					println(funcNode.Name.Name)
					return false
				}
				println(funcNode.Name.Name)
				if len(funcNode.Body.List) > 0 {
					if bb, ok := funcNode.Body.List[0].(*ast.ReturnStmt); ok {
						if _, okk := bb.Results[0].(*ast.FuncLit); okk {
							funcNode.Body = bb.Results[0].(*ast.FuncLit).Body
						}
					}
				}

				bodyCode := readCode(f, funcNode.Body)
				if strings.Contains(bodyCode, "grpc.Dial") {
					grpcConnVarName := ""
					checkOK := false
					for _, v := range funcNode.Body.List {
						switch v.(type) {
						case *ast.AssignStmt:
							if strings.Contains(readCode(f, v), "grpc.Dial") {
								name := v.(*ast.AssignStmt).Lhs[0].(*ast.Ident).Name
								grpcConnVarName = name
							}
						case *ast.DeferStmt:
							if grpcConnVarName != "" {
								fun := v.(*ast.DeferStmt).Call.Fun
								switch fun.(
								type) {
								case *ast.SelectorExpr:
									if ident, ok := fun.(*ast.SelectorExpr).X.(*ast.Ident); ok {
										if ident.Name == grpcConnVarName && fun.(*ast.SelectorExpr).Sel.Name == "Close" {
											checkOK = true
										}
									}
								case *ast.Ident:
								}
							}

						}
					}
					if !checkOK {
						println("not pass:")
						println(bodyCode)
					}
				}
			}
			return true
		})
	}

}
