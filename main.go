package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	//"golang.org/x/tools/go/types/typeutil"
	"sort"
)

func main() {

	inLoc := &TypeLocation{
		PackageName: "corego/service/xyt/view",
		TypeName:    "BaseView",
	}
	outLoc := &TypeLocation{
		PackageName: "corego/service/zhike-teacher/legacyapi",
		TypeName:    "GetTaskListResp",
	}

	api := NewAPI("test", "测试API", "/zhike/test", inLoc, outLoc)
	if err := api.Gen("corego/service/xyt/router"); err != nil {
		log.Fatal(err)
	}
	api.Print()

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

func getObjectFields(info *TypeLocation, packageTypesMap map[string]map[string]string) ([]string, error) {
	body := packageTypesMap[info.PackageName][info.TypeName]
	prefix := fmt.Sprintf("type %s struct{", info.TypeName)
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
	Name       string `json:"name"`
	JSONTag    string `json:"json_tag"`
	Comment    string `json:"comment"`
	Type       string `json:"type"`
	IsRepeated bool   `json:"is_repeated"`
	IsRef      bool   `json:"is_ref"`
	//IsMap      bool  暂不支持Map
}

func cleanRepeatedOrRef(filedTypeName string) string {
	s := strings.Replace(filedTypeName, "[]", "", -1)
	s = strings.Replace(s, "*", "", -1)
	return s
}

func getFieldNameFromType(filedType string) string {
	// type
	// []*corego/service/xyt/view.FooBar
	// []*FooBar

	if strings.Contains(filedType, ".") {
		return strings.Split(filedType, ".")[1]
	}

	return cleanRepeatedOrRef(filedType)

}

func getObjectField(structPkgInfo *TypeLocation, def string) (*ObjectField, error) {
	// name type tag
	// type tag

	// name type
	// FooBar  []*corego/service/xyt/view.FooBar
	// FooBars []*FooBar
	// String  string

	// type
	// []*corego/service/xyt/view.FooBar
	// []*FooBar

	var fieldName, fieldType string
	// remove tag
	tagStartIndex := strings.Index(def, "\"")
	if tagStartIndex > 0 {
		def = strings.TrimSpace(def[:tagStartIndex])
	}

	cols := strings.Split(def, " ")

	if len(cols) == 1 {
		// convert case 'type' to 'name type'
		cols = append([]string{getFieldNameFromType(cols[0])}, cols[0])
	}

	if len(cols) != 2 {
		return nil, fmt.Errorf("unsupport syntax : %s", def)
	}

	objectField := new(ObjectField)

	fieldName = cols[0]
	fieldType = cols[1]

	objectField.Name = fieldName

	if !isBuiltinType(fieldType) {
		objectField.IsRef = true
	}
	if strings.HasPrefix(fieldType, "[]") {
		objectField.IsRepeated = true
	}

	fieldBaseType := cleanRepeatedOrRef(fieldType)

	typePkgInfo := new(TypeLocation)

	pkgAndType := strings.Split(fieldBaseType, ".")

	if objectField.IsRef {
		if len(pkgAndType) == 1 {
			typePkgInfo.PackageName = structPkgInfo.PackageName
			typePkgInfo.TypeName = fieldBaseType
		} else if len(pkgAndType) == 2 {
			typePkgInfo.PackageName = cleanRepeatedOrRef(pkgAndType[0])
			typePkgInfo.TypeName = pkgAndType[1]
		} else {
			return nil, fmt.Errorf("unsupport syntax : %s", def)
		}
		objectField.Type = typePkgInfo.String()
	} else {
		objectField.Type = fieldBaseType
	}

	return objectField, nil
}

func isBuiltinType(t string) bool {
	builtinTypees := map[string]bool{
		"string":      true,
		"bool":        true,
		"byte":        true,
		"int":         true,
		"int32":       true,
		"int64":       true,
		"uint":        true,
		"uint32":      true,
		"uint64":      true,
		"float":       true,
		"float32":     true,
		"float64":     true,
		"interface{}": true,
	}
	return builtinTypees[t]
}

type Object struct {
	ID     string         `json:"id"`
	Fields []*ObjectField `json:"fields"`
}

type API struct {
	Name            string             `json:"name"`
	Comment         string             `json:"comment"`
	RouterPath      string             `json:"router_path"`
	InArgument      *Object            `json:"in_argument"`
	OutArgument     *Object            `json:"out_argument"`
	ObjectsMap      map[string]*Object `json:"objects_map"`
	inArgumentLoc   *TypeLocation
	outArgumentLoc  *TypeLocation
	packageTypesMap map[string]map[string]string
	debug           bool
}

func NewAPI(name string, comment string, routerPath string, inArgumentLoc, outArgumentLoc *TypeLocation) *API {
	return &API{Name: name, Comment: comment, RouterPath: routerPath, inArgumentLoc: inArgumentLoc, outArgumentLoc: outArgumentLoc}
}

func (api *API) getTypeInfo(query *TypeLocation, rootObj *Object, dep int) error {
	// println(query.PackageName, query.TypeName)
	fields, err := getObjectFields(query, api.packageTypesMap)
	if err != nil {
		return err
	}

	rootObj.ID = query.String()
	rootObj.Fields = make([]*ObjectField, 0)
	for _, v := range fields {
		t, err := getObjectField(query, v)
		prefixSpace := ""
		for i := 0; i < dep; i++ {
			prefixSpace += "\t\t"
		}
		if err != nil {
			if api.debug {
				fmt.Printf("%s- %v\n", prefixSpace, err)
			}
			return err
		}
		if !strings.HasPrefix(t.Name, "XXX_") {
			rootObj.Fields = append(rootObj.Fields, t)
			if api.debug {
				fmt.Printf("%s- %+v\n", prefixSpace, t)
			}
			if t.IsRef {
				if err = api.getTypeInfo(newTypeLocation(t.Type), new(Object), dep+1); err != nil {
					return err
				}
			}
		}
	}
	api.ObjectsMap[rootObj.ID] = rootObj
	return nil
}

func (api *API) Gen(rootPackage string) error {
	pkgTypesMap, err := GetPackageTypesMap(rootPackage)
	if err != nil {
		return err
	}
	api.packageTypesMap = pkgTypesMap
	api.InArgument = new(Object)
	api.OutArgument = new(Object)
	api.ObjectsMap = map[string]*Object{}
	err = api.getTypeInfo(api.inArgumentLoc, api.InArgument, 0)
	if err != nil {
		return err
	}
	err = api.getTypeInfo(api.outArgumentLoc, api.OutArgument, 0)
	if err != nil {
		return err
	}
	// set json tag and comment
	astPkgCacheMap := map[string]map[string]*ast.Package{}
	for _, obj := range api.ObjectsMap {
		err = api.setObjectJSONTagAndComment(obj, astPkgCacheMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func (api *API) setObjectJSONTagAndComment(obj *Object, astPkgCacheMap map[string]map[string]*ast.Package) error {
	goPath := os.Getenv("GOPATH")
	t := newTypeLocation(obj.ID)
	var f map[string]*ast.Package
	var err error
	if _, ok := astPkgCacheMap[t.PackageName]; !ok {
		fset := token.NewFileSet()
		f, err = parser.ParseDir(fset, filepath.Join(goPath, "src", t.PackageName), nil, parser.ParseComments)
		if err != nil {
			return err
		}
	} else {
		f = astPkgCacheMap[t.PackageName]
	}

	for _, v := range f {
		i, err := findStructInfo(t.TypeName, v)
		if err != nil {
			return err
		}
		for k, field := range obj.Fields {
			field.Comment = i.Fields[k].DocComment
			field.JSONTag = i.Fields[k].JSONTag
		}
	}
	return nil
}

func (api *API) Print() {
	b, _ := json.MarshalIndent(api, "", "\t")
	fmt.Println(string(b))
}
