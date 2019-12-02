package docspace

// DocGenerator receive a DocGenContext output a doc
type DocGenerator interface {
	// Generate doc
	Gen(ctx *DocGenContext) (output []byte, err error)
	// Generator Name
	Name() string
}

type DocGenContext struct {
	Tag     string
	APIs    []*API
	Config  Config
}
