package mkdoc

import (
	"log"
)

// DocGenerator receive a DocGenContext output a doc
type DocGenerator interface {
	// Generate doc
	Gen(ctx *DocGenContext) (output *GeneratedOutput, err error)
	// Generator Name
	Name() string
}

type DocGenContext struct {
	Tag    string
	APIs   []*API
	Config Config
	RefObj map[LangObjectId]*Object
	Args   map[string]string
}

var generators map[string]DocGenerator

// RegisterGenerator to global generator
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

// GetGenerators get all registered generator
func GetGenerators() map[string]DocGenerator {
	return generators
}

type GeneratedOutput struct {
	Files []*GeneratedFile
}

type GeneratedFile struct {
	Name string
	Data []byte
}
