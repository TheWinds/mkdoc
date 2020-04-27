package objmock

import "github.com/thewinds/mkdoc"

type ObjectMocker interface {
	Mock(object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error)
}
