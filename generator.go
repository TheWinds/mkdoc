package docspace

import "log"

// DocGenerator receive a DocGenContext output a doc
type DocGenerator interface {
	// Generate doc
	Gen(ctx *DocGenContext) (output []byte, err error)
	// Generator Name
	Name() string
	// File ext
	FileExt() string
}

type DocGenContext struct {
	Tag    string
	APIs   []*API
	Config Config
}

var generators map[string]DocGenerator

// RegisterGenerator to global generators
func RegisterGenerator(generator DocGenerator) {
	if generators == nil {
		generators = make(map[string]DocGenerator)
	}
	name := generator.Name()
	if generators[name] != nil {
		log.Fatalf("duplicate register generator : %s", name)
	}
	generators[name] = generator
}

// GetGenerators get all registered generators
func GetGenerators() map[string]DocGenerator {
	return generators
}
