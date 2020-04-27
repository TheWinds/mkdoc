package mkdoc

import (
	"github.com/thewinds/mkdoc/schema"
)

// API def
type API struct {
	schema.API
	InArgument  *Object `json:"in_argument"`
	OutArgument *Object `json:"out_argument"`
	Mime        *MimeType
}

// Build 生成API信息
func (api *API) Build() error {
	if api.Mime.In == "" {
		api.Mime.In = GetProject().Config.Mime.In
	}
	if api.Mime.Out == "" {
		api.Mime.Out = GetProject().Config.Mime.Out
	}
	return nil
	//return api.LinkBaseType()
}

//func (api *API) LinkBaseType() error {
//	project := GetProject()
//	if project.Config.BaseType == "" {
//		return nil
//	}
//	for _, d := range api.Disables {
//		if d == "base_type" {
//			return nil
//		}
//	}
//
//	baseTypObj := project.GetObject(project.Config.BaseType)
//	if baseTypObj == nil {
//		return fmt.Errorf("link base type: base type (%s) not found", project.Config.BaseType)
//	}
//	// clone
//	baseTypObj = baseTypObj.Clone()
//	project.AddObject(baseTypObj.ID, baseTypObj)
//
//	if api.OutArgument == nil {
//		api.OutArgument = baseTypObj
//		return nil
//	}
//
//	var tField, arrayTField *ObjectField
//	for _, v := range baseTypObj.Fields {
//		docTag := v.Tag.GetValue("doc")
//		switch {
//		case docTag == "T" && v.Type.Name == "interface{}":
//			tField = v
//		case docTag == "[]T" && v.Type.Name == "interface{}":
//			arrayTField = v
//		}
//	}
//	// object        => T
//	// array object  => []T,T
//	var toSelect []*ObjectField
//	if api.OutArgument.Type.IsRepeated {
//		toSelect = []*ObjectField{arrayTField, tField}
//	} else {
//		toSelect = []*ObjectField{tField}
//	}
//
//	var selected *ObjectField
//	for _, field := range toSelect {
//		if field != nil {
//			selected = field
//			break
//		}
//	}
//	if selected == nil {
//		return fmt.Errorf("link base type: please set `doc:\"T\"` or `doc:\"[]T\"` on generic(interface{}) field")
//	}
//	selected.Type.Name = "object"
//	selected.Type.Ref = api.OutArgument.ID
//	api.OutArgument = baseTypObj
//	return nil
//}
