package mkdoc

import (
	"fmt"
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
	DocLocation    string        `json:"doc_location"`
	Disables       []string      `json:"disables"`
	Annotation     DocAnnotation `json:"annotation"`
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
	return api.LinkBaseType()
}

func (api *API) LinkBaseType() error {
	project := GetProject()
	if project.Config.BaseType == "" {
		return nil
	}
	for _, d := range api.Disables {
		if d == "base_type" {
			return nil
		}
	}

	baseTyp := new(Object)
	err := api.getObjectInfoV2(newTypeLocation(project.Config.BaseType), baseTyp, 0)
	if err != nil {
		return err
	}
	var repeated bool
	if api.OutArgumentLoc != nil {
		repeated = api.OutArgumentLoc.IsRepeated
	}

	var linkFieldName string
	for _, v := range baseTyp.Fields {
		if v.DocTag == "[]T" && repeated {
			linkFieldName = v.Name
			break
		}
		if v.DocTag == "T" {
			linkFieldName = v.Name
		}
	}
	if linkFieldName == "" {
		return nil
	}
	if api.OutArgument == nil {
		api.OutArgument = baseTyp
		api.OutArgumentLoc = nil
		return nil
	}

	err = api.linkField2Object(baseTyp, linkFieldName, api.OutArgument.ID, repeated)
	if err != nil {
		return err
	}
	api.OutArgument = baseTyp
	api.OutArgumentLoc = nil
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
			if fromField.BaseType == "interface{}" {
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
	project := GetProject()
	var structInfo *GoStructInfo
	var err error

	if project.Config.UseGOModule {
		pkgAbsPath := strings.Replace(query.PackageName, project.ModulePkg, project.ModulePath, 1)
		structInfo, err = new(StructFinder).Find(pkgAbsPath, query.TypeName)
		if err != nil {
			return err
		}
		if structInfo == nil {
			return fmt.Errorf("struct %s not found\n", query)
		}
	} else {
		goSrcPaths := GetGOSrcPaths()
		pkgAbsPaths := make([]string, 0)
		for _, p := range goSrcPaths {
			pkgAbsPath := filepath.Join(p, query.PackageName)
			pkgAbsPaths = append(pkgAbsPaths, pkgAbsPath)
			if _, err := os.Stat(pkgAbsPath); err != nil {
				continue
			}
			structInfo, err = new(StructFinder).Find(pkgAbsPath, query.TypeName)
			if err != nil && err != errGoStructNotFound {
				return err
			}
			if structInfo != nil {
				break
			}
		}
		if structInfo == nil {
			return fmt.Errorf("struct %s not found in any of:\n	%s", query, strings.Join(pkgAbsPaths, "\n	"))
		}
	}

	rootObj.ID = query.String()
	rootObj.Fields = make([]*ObjectField, 0)

	for _, field := range structInfo.Fields {
		// TODO: filter by encoder
		if field.JSONTag == "-" {
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
			XMLTag:     field.XMLTag,
			DocTag:     field.DocTag,
			Comment:    comment,
			Type:       field.GoType.Location().String(),
			BaseType:   field.GoType.Name,
			IsRepeated: field.GoType.IsRep,
			IsRef:      field.GoType.IsRef,
		}
		rootObj.Fields = append(rootObj.Fields, objField)
		if objField.IsRef && api.ObjectsMap[rootObj.ID] == nil {
			if rootObj.ID == objField.Type {
				continue
			}
			if err := api.getObjectInfoV2(field.GoType.Location(), new(Object), dep+1); err != nil {
				return err
			}
		}
	}
	api.ObjectsMap[rootObj.ID] = rootObj
	return nil
}
