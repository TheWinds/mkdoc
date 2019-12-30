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

func BuiltinObjects() []*Object {

	var objects []*Object
	for _, t := range builtinTypes {
		objects = append(objects, &Object{
			ID:     t,
			Type:   &ObjectType{Name: t},
			Loaded: true,
		})
	}
	return objects
}
