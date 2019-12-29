package mkdoc

import (
	"fmt"
	"strings"
)

// ObjectField filed info
type ObjectField struct {
	Name string
	Desc string
	Type *ObjectType
	Tag  *ObjectFieldTag
}

type ObjectFieldTag struct {
	raw string
	m   map[string]*tagNameOptions
}

func (o *ObjectFieldTag) parse() error {
	raw := o.raw
	if len(raw) == 0 {
		return nil
	}
	if raw[0] == '`' && raw[len(raw)-1] == '`' {
		raw = raw[1 : len(raw)-1]
	}
	tags := strings.Fields(raw)
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

func mustObjectFieldTag(raw string) *ObjectFieldTag {
	tag := &ObjectFieldTag{raw: raw}
	err := tag.parse()
	if err != nil {
		panic(err)
	}
	return tag
}

