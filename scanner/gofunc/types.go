package gofunc

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
