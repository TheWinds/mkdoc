package main

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
	"sort"
	"strings"
)


type TypeLocation struct {
	PackageName string
	TypeName    string
}

func (t *TypeLocation) String() string {
	return fmt.Sprintf("%s.%s", t.PackageName, t.TypeName)
}

func newTypeLocation(raw string) *TypeLocation {
	t := &TypeLocation{}
	if strings.HasPrefix(raw, "*") {
		raw = raw[1:]
	} else if strings.HasPrefix(raw, "[]*") {
		raw = raw[3:]
	} else if strings.HasPrefix(raw, "[]") {
		raw = raw[2:]
	}
	e := strings.Split(raw, ".")
	t.PackageName, t.TypeName = e[0], e[1]
	return t
}



func GetPackageTypesMap(root string) (map[string]map[string]string, error) {
	cfg := &packages.Config{
		Mode: packages.LoadTypes,
	}

	lpkgs, err := packages.Load(cfg, root)
	if err != nil {
		panic(err)
		return nil, err
	}

	// 遍历所有依赖包
	var all []*packages.Package // postorder
	seen := make(map[*packages.Package]bool)
	var visit func(*packages.Package)
	visit = func(lpkg *packages.Package) {
		if !seen[lpkg] {
			seen[lpkg] = true

			// visit imports
			var importPaths []string
			for path := range lpkg.Imports {
				importPaths = append(importPaths, path)
			}
			sort.Strings(importPaths) // for determinism
			for _, path := range importPaths {
				visit(lpkg.Imports[path])
			}

			all = append(all, lpkg)
		}
	}
	for _, lpkg := range lpkgs {
		visit(lpkg)
	}
	lpkgs = all

	pkgTypesMap := map[string]map[string]string{}

	// 提取类型信息
	for _, lpkg := range lpkgs {
		if lpkg.Types != nil {
			qual := types.RelativeTo(lpkg.Types)
			scope := lpkg.Types.Scope()
			typesMap := map[string]string{}
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				if !obj.Exported() {
					continue // skip unexported names
				}

				ts := types.ObjectString(obj, qual)
				if strings.Contains(ts, "type") && strings.Contains(ts, "struct") {
					// type Word struct{Word
					typesMap[obj.Name()] = ts
				}
			}
			pkgTypesMap[lpkg.PkgPath] = typesMap
		}
	}

	return pkgTypesMap, nil
}
