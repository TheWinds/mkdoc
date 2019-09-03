package docspace

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
	Name           string             `json:"name"`
	Desc           string             `json:"desc"`
	Path           string             `json:"path"`
	Method         string             `json:"method"` // post get delete patch ; query mutation
	Type           string             `json:"type"`   // echo_handle graphql
	Tags           []string           `json:"tags"`
	InArgument     *Object            `json:"in_argument"`
	OutArgument    *Object            `json:"out_argument"`
	ObjectsMap     map[string]*Object `json:"objects_map"`
	inArgumentLoc  *TypeLocation
	outArgumentLoc *TypeLocation
	debug          bool
}

func NewAPI(name string, comment string, routerPath string, inArgumentLoc, outArgumentLoc *TypeLocation) *API {
	return &API{Name: name, Desc: comment, Path: routerPath, inArgumentLoc: inArgumentLoc, outArgumentLoc: outArgumentLoc}
}

// Build 生成API信息
// 得到所有依赖类型的信息、字段JSONTag以及DocComment
func (api *API) Build() error {
	if api.InArgument == nil {
		api.InArgument = new(Object)
	}
	if api.OutArgument == nil {
		api.OutArgument = new(Object)
	}
	api.ObjectsMap = map[string]*Object{}
	err := api.getObjectInfo(api.inArgumentLoc, api.InArgument, 0)
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
	fmt.Println(api.MakeMarkdown())
}

func (api *API) MakeMarkdown() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("### %s\n", api.Name))
	if len(strings.TrimSpace(api.Desc)) > 0 {
		sb.WriteString(fmt.Sprintf("> %s\n", api.Desc))
	}
	sb.WriteString(fmt.Sprintf("- %s %s\n", api.Method, api.Type))
	sb.WriteString(fmt.Sprintf("```\n"))
	sb.WriteString(fmt.Sprintf("[path] %s\n", api.Path))
	sb.WriteString(fmt.Sprintf("```\n"))

	sb.WriteString("- 参数\n")
	sb.WriteString(fmt.Sprintf("```json\n"))
	if api.inArgumentLoc != nil && api.inArgumentLoc.IsRepeated {
		sb.WriteString(fmt.Sprintf("[\n"))
	}
	sb.WriteString(fmt.Sprintf("%s", api.JSON(api.InArgument)))
	if api.inArgumentLoc != nil && api.inArgumentLoc.IsRepeated {
		sb.WriteString(fmt.Sprintf("]\n"))
	} else {
		sb.WriteString(fmt.Sprintf("\n"))
	}
	sb.WriteString(fmt.Sprintf("```\n"))
	sb.WriteString("- 返回\n")

	sb.WriteString(fmt.Sprintf("```json\n"))
	if api.outArgumentLoc != nil && api.outArgumentLoc.IsRepeated {
		sb.WriteString(fmt.Sprintf("["))
	}
	sb.WriteString(fmt.Sprintf("%s", api.JSON(api.OutArgument)))
	if api.outArgumentLoc != nil && api.outArgumentLoc.IsRepeated {
		sb.WriteString(fmt.Sprintf("]\n"))
	} else {
		sb.WriteString(fmt.Sprintf("\n"))
	}
	sb.WriteString(fmt.Sprintf("```\n"))
	return sb.String()
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
				api.writeJSONToken(MockField(field.Type, field.JSONTag), 0, sb)
				if i != len(obj.Fields)-1 {
					api.writeJSONToken(",", 0, sb)
				}
				api.writeJSONToken("\t# "+strings.TrimSuffix(field.Comment, "\n")+"\n", 0, sb)
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
	if query == nil {
		return nil
	}
	// println(query.PackageName, query.TypeName)
	fields, err := api.getObjectFields(query)
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

func (api *API) getObjectFields(info *TypeLocation) ([]string, error) {
	typesMap, err := GetPackageTypesMap(info.PackageName)
	if err != nil {
		return nil, err
	}
	body := typesMap[info.TypeName]
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
