package schema

type Schema struct {
	APIs    []*API    `json:"apis"`
	Objects []*Object `json:"objects"`
}
