package main

import (
	"fmt"
	"strings"
)

// ObjectField 字段
type ObjectField struct {
	Name       string `json:"name"`
	JSONTag    string `json:"json_tag"`
	Comment    string `json:"comment"`
	Type       string `json:"type"`
	IsRepeated bool   `json:"is_repeated"`
	IsRef      bool   `json:"is_ref"`
	//IsMap      bool  暂不支持Map
}

// 清除数组或指针前缀
func cleanRepeatedOrRef(filedTypeName string) string {
	s := strings.Replace(filedTypeName, "[]", "", -1)
	s = strings.Replace(s, "*", "", -1)
	return s
}

// 从go-type定义中获取到字段信息
func getFieldNameFromType(filedType string) string {
	// type
	// []*corego/service/xyt/view.FooBar
	// []*FooBar

	if strings.Contains(filedType, ".") {
		return strings.Split(filedType, ".")[1]
	}

	return cleanRepeatedOrRef(filedType)

}

// getObjectField
// 从go-type定义中获取到object信息
// 目前不支持 map
func getObjectField(structPkgInfo *TypeLocation, def string) (*ObjectField, error) {
	// name type tag
	// type tag
	// name type
	// FooBar  []*corego/service/xyt/view.FooBar
	// FooBars []*FooBar
	// String  string
	// Strings  []string

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

	if !isBuiltinType(cleanRepeatedOrRef(fieldType)) {
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
