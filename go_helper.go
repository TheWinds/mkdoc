package mkdoc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type GoModuleInfo struct {
	ModulePkg  string
	ModulePath string
}

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

// GetScanDirs get the dirs to scan
func GetScanDirs(pkg string, mod bool, filter func(dirName string) bool) []string {
	var dirs []string
	var roots []string

	if mod {
		roots = append(roots, pkg)
	} else {
		for _, srcPath := range GetGOSrcPaths() {
			roots = append(roots, filepath.Join(srcPath, pkg))
		}
	}

	for _, root := range roots {
		subDirs := GetSubDirs(root)
		// filter path
		for _, dir := range subDirs {
			if filter == nil || filter(dir) {
				dirs = append(dirs, dir)
			}
		}
	}
	return dirs
}

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

var importCache sync.Map

// GetFileImportsAtNode infer filename from node and then get the file imports
func GetFileImportsAtNode(node ast.Node, pkg *ast.Package, fileset *token.FileSet, mod *GoModuleInfo) map[string]string {
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
		fileImportMap[""] = getFilePkgPath(fileName, mod)
		m = fileImportMap
		importCache.Store(fileName, m)
	}
	return m.(map[string]string)
}

func GetFileImportsAtFile(fileName string, mod *GoModuleInfo) (map[string]string, error) {
	f := token.NewFileSet()
	file, err := parser.ParseFile(f, fileName, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	m, ok := importCache.Load(fileName)
	if !ok {
		fileImportMap := make(map[string]string)
		for _, v := range file.Imports {
			importName := ""
			importPath := strings.Replace(v.Path.Value, "\"", "", -1)
			if v.Name != nil {
				importName = v.Name.Name
			} else {
				importName = filepath.Base(importPath)
			}
			fileImportMap[importName] = importPath
		}
		fileImportMap[""] = getFilePkgPath(fileName, mod)
		m = fileImportMap
		importCache.Store(fileName, m)
	}
	return m.(map[string]string), nil
}

// getFilePkgPath get go package name from absolute file name
func getFilePkgPath(fileName string, mod *GoModuleInfo) string {
	if mod == nil {
		goSrcPaths := GetGOSrcPaths()
		for _, v := range goSrcPaths {
			if strings.HasPrefix(fileName, v) {
				return filepath.Dir(fileName[len(v):])
			}
		}
		return ""
	}
	rel := strings.Replace(fileName, mod.ModulePath, "", 1)
	rel = filepath.Dir(rel)
	rel = strings.TrimRight(rel, string(os.PathSeparator))
	return filepath.Join(mod.ModulePkg, rel)
}

var (
	slashSlash = []byte("//")
	moduleStr  = []byte("module")
)

// ModulePath returns the module path from the gomod file text.
// If it cannot find a module path, it returns an empty string.
// It is tolerant of unrelated problems in the go.mod file.
// Copy from go sdk 1.12.7
func ModulePath(mod []byte) string {
	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}
	return "" // missing module path
}

// FindGOModAbsPath find the first(in dep) absolute path which contains go.mod file
func FindGOModAbsPath(root string) string {
	absPath := ""
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && info.Name() == "go.mod" {
			absPath, _ = filepath.Abs(filepath.Dir(path))
			return filepath.SkipDir
		}
		return nil
	})
	return absPath
}

var (
	globalFileset       = token.NewFileSet()
	globalDirParseCache = make(map[string]map[string]*ast.Package)
)

// ParseDir parser dir and add file to global fileset
// also cached parsed packages
func ParseDir(dir string) (map[string]*ast.Package, *token.FileSet, error) {
	if globalDirParseCache[dir] == nil {
		pkgs, err := parser.ParseDir(globalFileset, dir, nil, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}
		globalDirParseCache[dir] = pkgs
	}
	return globalDirParseCache[dir], globalFileset, nil
}

func CheckGoScanPath(pkg string, enableGoMod bool) error {
	// check if the pkg to scan is exist
	if pkg == "" {
		return fmt.Errorf("please config a pkg to scan in conf.yaml")
	}

	if enableGoMod {
		path := pkg
		if !filepath.IsAbs(path) {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			path = filepath.Join(wd, path)
		}
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("no such file or directory: %s\n", path)
		}
	} else {
		goPaths := GetGOPaths()
		pkgExist := false
		for _, gopath := range goPaths {
			if _, err := os.Stat(filepath.Join(gopath, "src", pkg)); err == nil {
				pkgExist = true
			}
		}
		if !pkgExist {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("error: package \"%s\" is not found in any of:\n", pkg))
			for _, gopath := range goPaths {
				sb.WriteString(fmt.Sprintln("  ", filepath.Join(gopath, "src", pkg)))
			}
			return fmt.Errorf("%s", sb.String())
		}
	}
	return nil
}
