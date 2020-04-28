package mkdoc

import (
	"log"
)

type ObjectLoader interface {
	Load(ts TypeScope) (*Object, error)
	LoadAll(tss []TypeScope) ([]*Object, error)
	Add(object *Object) error
	GetObjectId(ts TypeScope) (string, error)
	SetConfig(cfg *ObjectLoaderConfig)
	Lang() string
}

type TypeScope struct {
	FileName string
	TypeName string
}

type ObjectLoaderConfig struct {
	Config
}

var objectLoaders map[string]ObjectLoader

// RegisterObjectLoader to global object loader
func RegisterObjectLoader(loader ObjectLoader) {
	if objectLoaders == nil {
		objectLoaders = make(map[string]ObjectLoader)
	}
	lang := loader.Lang()
	if objectLoaders[lang] != nil {
		log.Fatalf("duplicate register object loader : %s", lang)
	}
	objectLoaders[lang] = loader
}

// GetObjectLoaders get all registered object loaders
func GetObjectLoaders() map[string]ObjectLoader {
	return objectLoaders
}

// GetObjectLoader get one registered object loader
func GetObjectLoader(lang string) ObjectLoader {
	return objectLoaders[lang]
}
