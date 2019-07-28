package main

import (
	"errors"
	"fmt"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"

	//"golang.org/x/tools/go/types/typeutil"
	"sort"
)

func main() {

	pkgTypesMap, err := GetPackageTypesMap("corego/service/xyt/api")
	if err != nil {
		log.Fatal(err)
	}
	query := &typePkgInfo{
		Pkg:  "corego/service/xyt/view",
		Name: "UserWorkListView",
	}
	GetTypeInfo(query, pkgTypesMap)

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
		//println(lpkg.PkgPath, lpkg.Name)
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

	//for k, v := range pkgTypesMap {
	//	println("pkg:", k)
	//	for _, vv := range v {
	//		fmt.Printf("\t%s\n", vv)
	//	}
	//}

	return pkgTypesMap, nil
}

// queryType eg. "corego/service/xyt/api.ABC"
func GetTypeInfo(query *typePkgInfo, packageTypesMap map[string]map[string]string) {
	println(query.Pkg, query.Name)
	fields, err := getObjectFields(query, packageTypesMap)
	if err != nil {
		log.Fatal(err)
		return
	}
	for k, v := range fields {
		fmt.Println(k, v)
	}
}

type typePkgInfo struct {
	Pkg  string
	Name string
}

func (t *typePkgInfo) String() string {
	return fmt.Sprintf("%s.%s", t.Pkg, t.Name)
}

func newTypePkgInfo(raw string) *typePkgInfo {
	t := &typePkgInfo{}
	if strings.HasPrefix(raw, "*") {
		raw = raw[1:]
	} else if strings.HasPrefix(raw, "[]*") {
		raw = raw[3:]
	} else if strings.HasPrefix(raw, "[]") {
		raw = raw[2:]
	}
	e := strings.Split(raw, ",")
	t.Pkg, t.Name = e[0], e[1]
	return t
}

func getObjectFields(info *typePkgInfo, packageTypesMap map[string]map[string]string) ([]string, error) {
	body := packageTypesMap[info.Pkg][info.Name]
	prefix := fmt.Sprintf("type %s struct{", info.Name)
	body = strings.Replace(body, prefix, "", 1)
	body = body[:len(body)-1]

	fields := strings.Split(body, ";")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields, nil
	// a.b -> c.d
	// a.c -> f.e
}

type ObjectField struct {
	Name       string
	Type       string
	IsRepeated bool
	//IsMap      bool  暂不支持Map
	IsRef bool
}

func newObjectField(info *typePkgInfo, def string) (*ObjectField, error) {
	objectField := &ObjectField{}
	cols := strings.Split(def, " ")
	hasTag := strings.Contains(cols[len(cols)-1], "\"")
	// name type [tag]
	// type [tag]
	// name type
	// type

	if hasTag {
		cols = cols[:len(cols)-1]
	}

	var baseType string

	switch len(cols) {
	case 1:
		baseType = getBaseType(cols[0])
		if isSimpleType(baseType) {
			objectField.Type = baseType
		} else {
			objectField.IsRef = true
			if strings.Contains(baseType, ".") {
				// 不在同一个package
				objectField.Type = baseType

			} else {
				// 在同一个package
				i := &typePkgInfo{
					Pkg:  info.Pkg,
					Name: baseType,
				}
				objectField.Type = i.String()
				objectField.Name = baseType
			}
		}

	case 2:
		objectField.Name = cols[0]

		if strings.Contains(cols[1], "[]") {
			objectField.IsRepeated = true
		}

		if strings.Contains(cols[1], "map[") {
			return nil, errors.New("map field is not support")
		}

		baseType = getBaseType(cols[1])

	}
}

func getBaseType(t string) string {
	if strings.HasPrefix(t, "[]") {
		t = t[2:]
	}

	if strings.HasPrefix(t, "*") {
		t = t[1:]
	}
	return t
}

func isSimpleType(t string) bool {
	simpleTypes := map[string]bool{
		"string":  true,
		"bool":    true,
		"int":     true,
		"int32":   true,
		"int64":   true,
		"uint":    true,
		"uint32":  true,
		"uint64":  true,
		"float":   true,
		"float32": true,
		"float64": true,
	}
	return simpleTypes[t]
}

type Object struct {
	ID     string
	Fields []*ObjectField
}
