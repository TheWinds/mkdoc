package mkdoc

import (
	"encoding/json"
	"github.com/thewinds/mkdoc/schema"
)

type Extension interface {
	Name() string
	Parse(schema *schema.Extension) (Extension, error)
}

type ExtensionGoTag struct {
	Tag *ObjectFieldTag
}

func (e *ExtensionGoTag) Name() string {
	return "go_tag"
}

func (e *ExtensionGoTag) Parse(schema *schema.Extension) (Extension, error) {
	var tag string
	var err error
	if err = json.Unmarshal(schema.Data, &tag); err != nil {
		return nil, err
	}
	e.Tag, err = NewObjectFieldTag(tag)
	if err != nil {
		return nil, err
	}
	return e, nil
}

type ExtensionUnknown struct {
	OriginExtensionName string
	OriginData          json.RawMessage
}

func (e *ExtensionUnknown) Name() string {
	return "unknown_extension"
}

func (e *ExtensionUnknown) Parse(schema *schema.Extension) (Extension, error) {
	e.OriginData = schema.Data
	e.OriginExtensionName = schema.Name
	return e,nil
}
