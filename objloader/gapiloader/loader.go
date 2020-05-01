package gapiloader

import (
	"github.com/thewinds/mkdoc"
)

func init() {
	mkdoc.RegisterObjectLoader(new(Loader).init())
}

type Loader struct {
	config *mkdoc.ObjectLoaderConfig
	cached map[string]*mkdoc.Object
}

func (l *Loader) init() *Loader {
	l.cached = make(map[string]*mkdoc.Object)
	for _, object := range BuiltinObjects() {
		l.cached[object.ID] = object
	}
	return l
}

func (l *Loader) Load(ts mkdoc.TypeScope) (*mkdoc.Object, error) {
	return l.cached[ts.TypeName], nil
}

func (l *Loader) LoadAll(tss []mkdoc.TypeScope) ([]*mkdoc.Object, error) {
	var objs []*mkdoc.Object
	for _, ts := range tss {
		objs = append(objs, l.cached[ts.TypeName])
	}
	return objs, nil
}

func (l *Loader) Add(object *mkdoc.Object) error {
	l.cached[object.ID] = object
	return nil
}

func (l *Loader) GetObjectId(ts mkdoc.TypeScope) (string, error) {
	return ts.TypeName, nil
}

func (l *Loader) SetConfig(cfg *mkdoc.ObjectLoaderConfig) {
	l.config = cfg
}

func (l *Loader) Lang() string {
	return "gapi"
}
