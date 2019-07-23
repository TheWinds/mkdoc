package main

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"

	//"golang.org/x/tools/go/types/typeutil"
	"sort"
)

func main() {

	cfg := &packages.Config{
		Mode: packages.LoadTypes,
	}

	lpkgs, err := packages.Load(cfg, "corego/service/xyt")
	if err != nil {
		panic(err)
		return
	}

	// We can't use packages.All because
	// we need an ordered traversal.

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
	for _, lpkg := range lpkgs {
		println(lpkg.Name)
		if lpkg.Types != nil {
			qual := types.RelativeTo(lpkg.Types)
			scope := lpkg.Types.Scope()
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				if !obj.Exported() {
					continue // skip unexported names
				}
				ts := types.ObjectString(obj, qual)
				if strings.Contains(ts, "type") && strings.Contains(ts, "struct") {
					fmt.Printf("\t%s\n", ts)
				}
				//if _, ok := obj.(*types.TypeName); ok {
				//	for _, meth := range typeutil.IntuitiveMethodSet(obj.Type(), nil) {
				//		if !meth.Obj().Exported() {
				//			continue // skip unexported names
				//		}
				//		fmt.Printf("\t%s\n", types.SelectionString(meth, qual))
				//	}
				//}
			}
		}
	}

}
