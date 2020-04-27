package goloader

import "github.com/thewinds/mkdoc"

var builtinTypes = []string{
	"string",
	"bool",
	"byte",
	"int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64",
	"float", "float32", "float64",
	"interface{}"}

var builtinTypesMap map[string]bool

func init() {
	builtinTypesMap = make(map[string]bool)
	for _, typ := range builtinTypes {
		builtinTypesMap[typ] = true
	}
}

func isBuiltinType(t string) bool {
	return builtinTypesMap[t]
}

func BuiltinObjects() []*mkdoc.Object {
	var objects []*mkdoc.Object
	for _, t := range builtinTypes {
		objects = append(objects, &mkdoc.Object{
			ID:     t,
			Type:   &mkdoc.ObjectType{Name: t},
			Loaded: true,
		})
	}
	return objects
}
