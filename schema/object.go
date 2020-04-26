package schema

import "encoding/json"

type Object struct {
	ID         string         `json:"id"`
	Type       *ObjectType    `json:"type"`
	Fields     []*ObjectField `json:"fields"`
	Extensions []*Extension   `json:"extensions"`
}

type ObjectField struct {
	Name       string       `json:"name"`
	Desc       string       `json:"desc"`
	Type       *ObjectType  `json:"type"`
	Extensions []*Extension `json:"extensions"`
}

type ObjectType struct {
	Name       string `json:"name"`
	Ref        string `json:"ref"`
	IsRepeated bool   `json:"is_repeated"`
}

type Extension struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}
