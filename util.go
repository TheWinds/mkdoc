package docspace

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// GetGOPaths get all go paths
func GetGOPaths() []string {
	pathEnv := os.Getenv("GOPATH")
	paths := strings.Split(pathEnv, ":")
	for i := 0; i < len(paths); i++ {
		paths[i] = strings.TrimSpace(paths[i])
	}
	return paths
}

// GetGOSrcPaths get all go src paths
func GetGOSrcPaths() []string {
	var paths []string
	for _, goPath := range GetGOPaths() {
		goSrcPath := filepath.Join(goPath, "src") + "/"
		paths = append(paths, goSrcPath)
	}
	return paths
}

var importCache sync.Map

// GetFileImportsAtNode
// infer filename from node and then get the file imports
func GetFileImportsAtNode(node ast.Node, pkg *ast.Package, fileset *token.FileSet) map[string]string {
	fileName := fileset.File(node.Pos()).Name()
	m, ok := importCache.Load(fileName)
	if !ok {
		fileImportMap := make(map[string]string)
		for _, v := range pkg.Files[fileName].Imports {
			importName := ""
			importPath := strings.Replace(v.Path.Value, "\"", "", -1)
			if v.Name != nil {
				importName = v.Name.Name
			} else {
				importName = filepath.Base(importPath)
			}
			fileImportMap[importName] = importPath
		}
		fileImportMap[""] = GetFilePkgPath(fileName)
		m = fileImportMap
		importCache.Store(fileName, m)
	}
	return m.(map[string]string)
}

// GetFilePkgPath
// get go package name from absolute file name
func GetFilePkgPath(fileName string) string {
	goSrcPaths := GetGOSrcPaths()
	for _, v := range goSrcPaths {
		if strings.HasPrefix(fileName, v) {
			return filepath.Dir(fileName[len(v):])
		}
	}
	return ""
}
