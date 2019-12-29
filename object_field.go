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
	m   map[string]string
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
	o.m = make(map[string]string)
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
		o.m[id] = body[1 : len(body)-1]
	}
	return nil
}

func (o *ObjectFieldTag) GetValue(tagName string) string {
	if o == nil {
		return ""
	}
	return o.m[tagName]
}

func (o *ObjectFieldTag) GetFirstValue(tagName string, sep string) string {
	if o == nil {
		return ""
	}
	v := o.m[tagName]
	i := strings.Index(v, sep)
	if i != -1 {
		return v[:i]
	}
	return v
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
