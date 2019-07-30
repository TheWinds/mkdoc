package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

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

// Gen 生成API信息
// 得到所有依赖类型的信息、字段JSONTag以及DocComment
func (api *API) Gen(rootPackage string) error {
	pkgTypesMap, err := GetPackageTypesMap(rootPackage)
	if err != nil {
		return err
	}
	api.packageTypesMap = pkgTypesMap
	api.InArgument = new(Object)
	api.OutArgument = new(Object)
	api.ObjectsMap = map[string]*Object{}
	err = api.getObjectInfo(api.inArgumentLoc, api.InArgument, 0)
	if err != nil {
		return err
	}
	err = api.getObjectInfo(api.outArgumentLoc, api.OutArgument, 0)
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

func (api *API) Print() {
	b, _ := json.MarshalIndent(api, "", "\t")
	fmt.Println(string(b))
}

func (api *API) PrintJSON() {
	//sb := new(strings.Builder)
	//api.printJSON(api.InArgument, 0, sb)
	//m := map[string]interface{}{}
	//err := json.Unmarshal([]byte(sb.String()), m)
	//fmt.Println(err)
	//b, _ := json.MarshalIndent(m, "", "\t")
	//fmt.Println(string(b))
	//
	//sb = new(strings.Builder)
	//api.printJSON(api.OutArgument, 0, sb)
	//fmt.Println(sb.String())

	mIn := make(map[string]interface{})
	api.buildMap(api.InArgument, mIn)
	mOut := make(map[string]interface{})
	api.buildMap(api.OutArgument, mOut)

	bin, _ := json.MarshalIndent(mIn, "", "\t")
	bout, _ := json.MarshalIndent(mOut, "", "\t")

	fmt.Println(string(bin))
	fmt.Println(string(bout))

}

func (api *API) buildMap(obj *Object, rootMap map[string]interface{}) {
	for _, field := range obj.Fields {
		if field.IsRef {
			f := map[string]interface{}{}
			api.buildMap(api.ObjectsMap[field.Type], f)
			if field.IsRepeated {
				rootMap[field.JSONTag] = []map[string]interface{}{f}
			} else {
				rootMap[field.JSONTag] = f
			}
		} else {
			rootMap[field.JSONTag] = 123
		}
	}
}

func (api *API) printJSON(obj *Object, dep int, sb *strings.Builder) {
	api.writeJSONToken("{\n", dep, sb)
	for _, field := range obj.Fields {
		k := fmt.Sprintf("\t\"%s\" : ", field.JSONTag)
		api.writeJSONToken(k, dep, sb)
		if !field.IsRef {
			api.writeJSONToken("123", dep, sb)
			api.writeJSONToken(",\n", dep, sb)
			continue
		}
		if field.IsRepeated {
			api.writeJSONToken("\t[\n", dep, sb)
			api.printJSON(api.ObjectsMap[field.Type], dep+1, sb)
			api.writeJSONToken(",", dep+1, sb)
			api.writeJSONToken("\t]\n", dep, sb)
		} else {
			api.printJSON(api.ObjectsMap[field.Type], dep+1, sb)
		}
		api.writeJSONToken(",\n", dep, sb)
	}
	api.writeJSONToken("}\n", dep, sb)
}

func (api *API) writeJSONToken(s string, dep int, sb *strings.Builder) {
	prefix := ""
	for i := 0; i < dep; i++ {
		prefix += "\t"
	}
	sb.WriteString(prefix + s)
}

func (api *API) getObjectInfo(query *TypeLocation, rootObj *Object, dep int) error {
	// println(query.PackageName, query.TypeName)
	fields, err := api.getObjectFields(query, api.packageTypesMap)
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
				if err = api.getObjectInfo(newTypeLocation(t.Type), new(Object), dep+1); err != nil {
					return err
				}
			}
		}
	}
	api.ObjectsMap[rootObj.ID] = rootObj
	return nil
}

func (api *API) getObjectFields(info *TypeLocation, packageTypesMap map[string]map[string]string) ([]string, error) {
	body := packageTypesMap[info.PackageName][info.TypeName]
	prefix := fmt.Sprintf("type %s struct{", info.TypeName)
	body = strings.Replace(body, prefix, "", 1)
	body = body[:len(body)-1]

	fields := strings.Split(body, ";")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields, nil
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
		i, err := findGOStructInfo(t.TypeName, v)
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

// 添加API页面	CURD
// API 			CURD
// Field 		CURD
// types 搜索
// 类型连接

type APIGroup struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	APIList     []*API `json:"api_list"`
}
