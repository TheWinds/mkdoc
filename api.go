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

func (api *API) PrintMarkdown() {
	fmt.Printf("### %s\n", api.Comment)
	fmt.Printf("#### %s\n", api.Name)
	fmt.Printf("#### %s\n", api.RouterPath)
	fmt.Println("- 参数")
	fmt.Printf("```json\n")
	fmt.Printf("%s\n", api.JSON(api.InArgument))
	fmt.Printf("```\n")
	fmt.Println("- 返回")
	fmt.Printf("```json\n")
	fmt.Printf("%s\n", api.JSON(api.OutArgument))
	fmt.Printf("```\n")
}

func (api *API) LinkField2Field(fromObj *Object, fromFieldName string, toObj *Object, toFieldName string) error {
	fromFieldIndex := -1
	for k, fromField := range fromObj.Fields {
		if fromField.Name == fromFieldName {
			if fromField.Type == "interface{}" {
				fromFieldIndex = k
				break
			} else {
				return fmt.Errorf("link filed should from a interface{} filed but got %s", fromField.Type)
			}
		}
	}
	if fromFieldIndex == -1 {
		return fmt.Errorf("type %s is not constains field %s", fromObj.ID, fromFieldName)

	}

	for _, toField := range toObj.Fields {
		if toField.Name == toFieldName {
			if !toField.IsRef {
				return fmt.Errorf("filed must link to a pointer filed")
			}
			fromObj.Fields[fromFieldIndex].IsRepeated = toField.IsRepeated
			fromObj.Fields[fromFieldIndex].Type = toField.Type
			fromObj.Fields[fromFieldIndex].IsRef = true
			return nil
		}
	}

	return fmt.Errorf("type %s is not constains field %s", toObj.ID, toFieldName)
}

func (api *API) LinkField2Object(fromObj *Object, fromFieldName string, toObjID string, isRepeated bool) error {

	for _, fromField := range fromObj.Fields {
		if fromField.Name == fromFieldName {
			if fromField.Type == "interface{}" {
				fromField.IsRef = true
				fromField.Type = toObjID
				fromField.IsRepeated = isRepeated
				return nil
			} else {
				return fmt.Errorf("link filed should from a interface{} filed but got %s", fromField.Type)
			}
		}
	}
	return fmt.Errorf("type %s is not constains field %s", fromObj.ID, fromFieldName)
}

func (api *API) Print() {
	b, _ := json.MarshalIndent(api, "", "\t")
	fmt.Println(string(b))
}

func (api *API) PrintJSON(obj *Object) {
	sb := new(strings.Builder)
	api.printJSON(obj, 0, sb)
	fmt.Println(sb.String())
}

func (api *API) JSON(obj *Object) string {
	sb := new(strings.Builder)
	api.printJSON(obj, 0, sb)
	return sb.String()
}

func (api *API) printJSON(obj *Object, dep int, sb *strings.Builder) {
	api.writeJSONToken("{\n", 0, sb)
	for i, field := range obj.Fields {
		k := fmt.Sprintf("    \"%s\" : ", field.JSONTag)
		api.writeJSONToken(k, dep, sb)
		if !field.IsRef {
			if field.IsRepeated {
				api.writeJSONToken("[", 0, sb)
				api.writeJSONToken("1,2,3", 0, sb)
				api.writeJSONToken("]", 0, sb)
				if i != len(obj.Fields)-1 {
					api.writeJSONToken(",\n", 0, sb)
				} else {
					api.writeJSONToken("\n", 0, sb)
				}
			} else {
				api.writeJSONToken("123", 0, sb)
				if i != len(obj.Fields)-1 {
					api.writeJSONToken(",", 0, sb)
				}
				api.writeJSONToken(" # "+strings.TrimSuffix(field.Comment,"\n") +"\n", 0, sb)
			}
			continue
		}
		if field.IsRepeated {
			api.writeJSONToken("[", 0, sb)
			api.printJSON(api.ObjectsMap[field.Type], dep+1, sb)
			api.writeJSONToken("]", dep, sb)
		} else {
			api.printJSON(api.ObjectsMap[field.Type], dep+1, sb)
		}
		if i != len(obj.Fields)-1 {
			api.writeJSONToken(",\n", 0, sb)
		} else {
			api.writeJSONToken("\n", 0, sb)
		}
	}
	api.writeJSONToken("}", dep, sb)
}

func (api *API) writeJSONToken(s string, dep int, sb *strings.Builder) {
	prefix := ""
	for i := 0; i < dep; i++ {
		prefix += "    "
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
			if t.IsRef && dep >= 0 {
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
	//TODO:支持多个GOPATH
	goPath := os.Getenv("GOPATH")
	t := newTypeLocation(obj.ID)
	var f map[string]*ast.Package
	var err error
	if astPkgCacheMap != nil {
		if _, ok := astPkgCacheMap[t.PackageName]; !ok {
			fset := token.NewFileSet()
			f, err = parser.ParseDir(fset, filepath.Join(goPath, "src", t.PackageName), nil, parser.ParseComments)
			if err != nil {
				return err
			}
			astPkgCacheMap[t.PackageName] = f
		} else {
			f = astPkgCacheMap[t.PackageName]
		}
	} else {
		fset := token.NewFileSet()
		f, err = parser.ParseDir(fset, filepath.Join(goPath, "src", t.PackageName), nil, parser.ParseComments)
		if err != nil {
			return err
		}
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
