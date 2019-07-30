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
				if err = api.getTypeInfo(newTypeLocation(t.Type), new(Object), dep+1); err != nil {
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

// 添加API页面	CURD
// API 			CURD
// Field 		CURD
// types 搜索
// 类型连接