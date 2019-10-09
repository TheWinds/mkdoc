package docspace

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
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
	err := api.getObjectInfoV2(api.inArgumentLoc, api.InArgument, 0)
	if err != nil {
		return err
	}
	err = api.getObjectInfoV2(api.outArgumentLoc, api.OutArgument, 0)
	if err != nil {
		return err
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
	if api.Type == "graphql" {
		sb.WriteString(fmt.Sprintf("```\n"))
		sb.WriteString(api.GQL())
		sb.WriteString(fmt.Sprintf("\n```\n"))
	}
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

func (api *API) GQL() string {
	sb := new(strings.Builder)
	ind := strings.LastIndex(api.Path, ":")
	opName := api.Path[ind+1:]
	args := make([]string, 0)
	argsInner := make([]string, 0)
	for _, field := range api.InArgument.Fields {
		var gqlTyp string
		switch field.Type {
		case "string":
			gqlTyp = "String!"
		case "bool":
			gqlTyp = "Boolean!"
		case "int", "int32", "int64", "uint", "uint32", "uint64":
			gqlTyp = "Int!"
		case "float", "float32", "float64":
			gqlTyp = "Float!"
		}
		if field.IsRepeated {
			gqlTyp = "[" + gqlTyp + "]!"
		}
		args = append(args, fmt.Sprintf("$%s: %s", field.JSONTag, gqlTyp))
		argsInner = append(argsInner, fmt.Sprintf("%s: $%s", field.JSONTag, field.JSONTag))
	}
	bodykw := "body"
	if api.outArgumentLoc != nil && api.outArgumentLoc.IsRepeated {
		bodykw = "bodys"
	}
	ql := `%s %s(%s) {
		%s(%s) {
		  total
		  %s%s
		  errorCode
		  errorMsg
		  success
		}
	  }`
	sb.WriteString(
		fmt.Sprintf(
			ql,
			api.Method,
			opName,
			strings.Join(args, ","),
			opName,
			strings.Join(argsInner, ","),
			bodykw,
			api.GQLBody(),
		))

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

func (api *API) GQLBody() string {
	sb := new(strings.Builder)
	api.printGQLBody(api.OutArgument, 0, sb)
	return sb.String()
}

func (api *API) printGQLBody(obj *Object, dep int, sb *strings.Builder) {
	api.writeJSONToken("{\n", 0, sb)
	for _, field := range obj.Fields {
		k := fmt.Sprintf("		      %s", field.JSONTag)
		api.writeJSONToken(k, dep, sb)
		if !field.IsRef {
			if field.IsRepeated {
				api.writeJSONToken("\n", 0, sb)
			} else {
				api.writeJSONToken("\n", 0, sb)
			}
			continue
		}
		if field.IsRepeated {
			api.printGQLBody(api.ObjectsMap[field.Type], dep+1, sb)
		} else {
			api.printGQLBody(api.ObjectsMap[field.Type], dep+1, sb)
		}
		api.writeJSONToken(",\n", 0, sb)
	}
	api.writeJSONToken("		  }", dep, sb)
}

func (api *API) writeJSONToken(s string, dep int, sb *strings.Builder) {
	prefix := ""
	for i := 0; i < dep; i++ {
		prefix += "    "
	}
	sb.WriteString(prefix + s)
}

func (api *API) getObjectInfoV2(query *TypeLocation, rootObj *Object, dep int) error {
	if query == nil {
		return nil
	}
	var structInfo *GoStructInfo
	goPaths := GetGOPaths()
	pkgPaths := make([]string, 0)
	for _, goPath := range goPaths {
		f := token.NewFileSet()
		pkgPath := filepath.Join(goPath, "src", query.PackageName)
		pkgPaths = append(pkgPaths, pkgPath)
		pkgs, err := parser.ParseDir(f, pkgPath, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, pkg := range pkgs {
			structInfo, err = findGOStructInfo(query.TypeName, pkg, f)
			if err != nil && err != ErrGoStructNotFound {
				return err
			}
		}
	}

	if structInfo == nil {
		return fmt.Errorf("struct %s not found in any of:\n  %s", query, strings.Join(pkgPaths, "\n"))
	}

	rootObj.ID = query.String()
	rootObj.Fields = make([]*ObjectField, 0)

	for _, field := range structInfo.Fields {
		if strings.HasPrefix(field.Name, "XXX_") {
			continue
		}
		// priority use doc comment
		var comment string
		if field.DocComment != "" {
			comment = field.DocComment
		} else {
			comment = field.Comment
		}
		objField := &ObjectField{
			Name:       field.Name,
			JSONTag:    field.JSONTag,
			Comment:    comment,
			Type:       field.GoType.Location().String(),
			IsRepeated: field.GoType.IsRep,
			IsRef:      field.GoType.IsRef,
		}
		rootObj.Fields = append(rootObj.Fields, objField)
		if objField.IsRef && api.ObjectsMap[rootObj.ID] == nil {
			if err := api.getObjectInfoV2(field.GoType.Location(), new(Object), dep+1); err != nil {
				return err
			}
		}
	}
	api.ObjectsMap[rootObj.ID] = rootObj
	return nil
}
