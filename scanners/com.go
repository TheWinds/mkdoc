package scanners

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GetSubDirs(root string) []string {
	subDirs := make([]string, 0)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			subDirs = append(subDirs, path)
		}
		return nil
	})
	return subDirs
}

func PrintCode(f *token.FileSet, node ast.Node) {
	fmt.Println(ReadCode(f, node))
}

func ReadCode(f *token.FileSet, node ast.Node) string {
	ps := f.Position(node.Pos())
	pe := f.Position(node.End())
	file, _ := ioutil.ReadFile(ps.Filename)
	return string(file[ps.Offset:pe.Offset])
}
