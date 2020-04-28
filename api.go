package mkdoc

import (
	"github.com/thewinds/mkdoc/schema"
)

// API def
type API struct {
	schema.API
	InArgument  *Object `json:"in_argument"`
	OutArgument *Object `json:"out_argument"`
	Mime        *MimeType
}
