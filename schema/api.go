package schema

type API struct {
	Name           string            `json:"name"`
	Desc           string            `json:"desc"`
	Path           string            `json:"path"`
	Method         string            `json:"method"` // post get delete patch ; query mutation
	Type           string            `json:"type"`   // echo_handle graphql
	Tags           []string          `json:"tags"`
	Query          map[string]string `json:"query"`
	Header         map[string]string `json:"header"`
	InType         string            `json:"in_type"`
	OutType        string            `json:"out_type"`
	MimeIn         string            `json:"mime_in"`
	MimeOut        string            `json:"mime_out"`
	Source         string            `json:"src"`
	SourceFileName string            `json:"src_file_name"`
	SourceLineNum  int               `json:"src_line_num"`
	Language       string            `json:"lang"`
	Disables       []string          `json:"disables"`
}
