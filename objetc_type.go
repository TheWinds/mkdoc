package mkdoc

// ObjectType describe a object
//
// Name is one of below:
// object,
// string,
// bool,
// byte,
// interface{},
// int,int8,int16,int32,int64,
// uint,uint8,uint16,uint32,uint64,
// float,float32,float64
//
// Ref describe which object to reference
//
// IsRepeated will be true if that is a array/slice type
type ObjectType struct {
	Name       string
	Ref        string
	IsRepeated bool
}

func isBuiltinType(t string) bool {
	builtinTypes := map[string]bool{
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
	return builtinTypes[t]
}

func BuiltinObjects() []*Object {
	types := []string{
		"string",
		"bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float", "float32", "float64",
		"interface{}"}
	var objects []*Object
	for _, t := range types {
		objects = append(objects, &Object{
			ID:     t,
			Type:   &ObjectType{Name: t},
			Loaded: true,
		})
	}
	return objects
}
