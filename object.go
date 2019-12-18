package mkdoc

import (
	"fmt"
	"strings"
)

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

type ObjectFieldTag struct {
	raw string
	m   map[string]*tagNameOptions
}

func (o *ObjectFieldTag) parse() error {
	tags := strings.Fields(o.raw)
	o.m = make(map[string]*tagNameOptions)
	for _, t := range tags {
		r := strings.Split(t, ":")
		if len(r) != 2 {
			return fmt.Errorf("tag parse error:%v", o.raw)
		}
		id, body := r[0], r[1]
		id = strings.TrimSpace(id)
		body = strings.TrimSpace(body)
		if len(body) < 2 || !(body[0] == '"' && body[len(body)-1] == '"') {
			return fmt.Errorf("tag parse error:%v", o.raw)
		}
		nameOpts := strings.Split(body[1:len(body)-1], ",")
		tn := &tagNameOptions{}
		if len(nameOpts) > 1 {
			tn.Options = make(map[string]bool, len(nameOpts)-1)
		}
		for k, v := range nameOpts {
			if k == 0 {
				tn.Name = v
				continue
			}
			tn.Options[v] = true
		}
		o.m[id] = tn
	}
	return nil
}

func (o *ObjectFieldTag) GetTagName(tagID string) string {
	if o == nil {
		return ""
	}
	tn := o.m[tagID]
	if tn != nil {
		return tn.Name
	}
	return ""
}

func (o *ObjectFieldTag) HasTagOption(tagID string, optionName string) bool {
	if o == nil {
		return false
	}
	tn := o.m[tagID]
	if tn != nil {
		return tn.Options[optionName]
	}
	return false
}

type tagNameOptions struct {
	Name    string
	Options map[string]bool
}

func NewObjectFieldTag(raw string) (*ObjectFieldTag, error) {
	tag := &ObjectFieldTag{raw: raw}
	err := tag.parse()
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// ObjectField filed info
type ObjectField struct {
	Name string
	Desc string
	Type *ObjectType
	Tag  *ObjectFieldTag
}

// Object info
type Object struct {
	Type   *ObjectType
	Fields []*ObjectField
}
