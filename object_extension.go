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
	return e, nil
}

type GApiFieldExtension struct {
	Options struct {
		OmitEmpty   bool `json:"omit_empty"`
		RawData     bool `json:"raw_data"`
		FromHeader  bool `json:"from_header"`
		FromContext bool `json:"from_context"`
		FromParams  bool `json:"from_params"`
		Validate    bool `json:"validate"`
	}
}

func (e *GApiFieldExtension) Name() string {
	return "gapi_field"
}

func (e *GApiFieldExtension) Parse(schema *schema.Extension) (Extension, error) {
	if err := json.Unmarshal(schema.Data, &e.Options); err != nil {
		return nil, err
	}
	return e, nil
}
