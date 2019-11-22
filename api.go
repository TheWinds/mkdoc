package docspace

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// API def
type API struct {
	Name           string             `json:"name"`
	Desc           string             `json:"desc"`
	Path           string             `json:"path"`
	Method         string             `json:"method"` // post get delete patch ; query mutation
	Type           string             `json:"type"`   // echo_handle graphql
	Tags           []string           `json:"tags"`
	Query          map[string]string  `json:"query"`
	Header         map[string]string  `json:"header"`
	InArgument     *Object            `json:"in_argument"`
	OutArgument    *Object            `json:"out_argument"`
	InArgEncoder   string             `json:"in_arg_encoder"`
	OutArgEncoder  string             `json:"out_arg_encoder"`
	ObjectsMap     map[string]*Object `json:"objects_map"`
	InArgumentLoc  *TypeLocation
	OutArgumentLoc *TypeLocation
	DocLocation    string   `json:"doc_location"`
	Disables       []string `json:"disables"`
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
	err := api.getObjectInfoV2(api.InArgumentLoc, api.InArgument, 0)
	if err != nil {
		return err
	}
	err = api.getObjectInfoV2(api.OutArgumentLoc, api.OutArgument, 0)
	if err != nil {
		return err
	}
	return nil
}

func (api *API) linkField2Field(fromObj *Object, fromFieldName string, toObj *Object, toFieldName string) error {
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

func (api *API) linkField2Object(fromObj *Object, fromFieldName string, toObjID string, isRepeated bool) error {

	for _, fromField := range fromObj.Fields {
		if fromField.Name == fromFieldName {
			if fromField.Type == "interface{}" {
				fromField.IsRef = true
				fromField.Type = toObjID
				fromField.IsRepeated = isRepeated
				return nil
			}
			return fmt.Errorf("link filed should from a interface{} filed but got %s", fromField.Type)
		}
	}
	return fmt.Errorf("type %s is not constains field %s", fromObj.ID, fromFieldName)
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
		if _, err := os.Stat(pkgPath); err != nil {
			continue
		}
		pkgPaths = append(pkgPaths, pkgPath)
		pkgs, err := parser.ParseDir(f, pkgPath, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, pkg := range pkgs {
			structInfo, err = findGOStructInfo(query.TypeName, pkg, f)
			if err != nil && err != errGoStructNotFound {
				return err
			}
			if structInfo != nil {
				break
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
			BaseType:   field.GoType.Name,
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
