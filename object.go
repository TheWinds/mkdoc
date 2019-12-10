package mkdoc

// ObjectField filed info
type ObjectField struct {
	Name       string `json:"name"`
	JSONTag    string `json:"json_tag"`
	XMLTag     string `json:"xml_tag"`
	DocTag     string `json:"doc_tag"`
	Comment    string `json:"comment"`
	Type       string `json:"type"`
	BaseType   string `json:"baseType"`
	IsRepeated bool   `json:"is_repeated"`
	IsRef      bool   `json:"is_ref"`
	//IsMap      bool  暂不支持Map
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

// Object fields and id to describe a type
type Object struct {
	ID     string         `json:"id"`
	Fields []*ObjectField `json:"fields"`
}
