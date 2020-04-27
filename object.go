package mkdoc

import (
	"fmt"
	"math/rand"
)

// Object info
type Object struct {
	ID         string
	Type       *ObjectType
	Fields     []*ObjectField
	Extensions []Extension
	Loaded     bool
}

type LangObjectId struct {
	Lang string
	Id   string
}

// Clone object with a random id
func (obj *Object) Clone() *Object {
	newObj := new(Object)
	newObj.ID = randObjectID("clone")
	t := *obj.Type
	newObj.Type = &t
	for _, field := range obj.Fields {
		ft := *field.Type
		newField := &ObjectField{
			Name: field.Name,
			Desc: field.Desc,
			Type: &ft,
		}
		// TODO copy extensions
		//field.Extensions
		newObj.Fields = append(newObj.Fields, newField)
	}
	newObj.Loaded = obj.Loaded
	return newObj
}

func randObjectID(s string) string {
	return fmt.Sprintf("obj_%s_#%d", s, rand.Int63())
}

// CreateRootObject
// create root object by package and type
// returns the created obj(return Object[0]) an refs object
func CreateRootObject(pkgTyp string, loadFn func(id string) *Object) ([]*Object, error) {
	var i int
	for i = 0; i < len(pkgTyp); i += 2 {
		if pkgTyp[i] == '[' {
			if i+1 >= len(pkgTyp) || pkgTyp[i+1] != ']' {
				return nil, fmt.Errorf("invaild type '%s'", pkgTyp)
			}
			continue
		}
		break
	}
	arrDep := i / 2
	pkgTypPath := pkgTyp[i:]
	var leaf *Object
	if loadFn != nil {
		leaf = loadFn(pkgTypPath)
	}
	if leaf == nil {
		leaf = &Object{
			ID: pkgTypPath,
			Type: &ObjectType{
				Name:       "object",
				Ref:        "",
				IsRepeated: false,
			},
			Fields: nil,
			Loaded: false,
		}
	}
	return CreateArrayObject(leaf, arrDep), nil
}

// Create a n-dimensional(dep) array object
func CreateArrayObject(leaf *Object, dep int) []*Object {
	var deps []*Object
	root := leaf
	deps = append(deps, root)
	for k := 0; k < dep; k++ {
		obj := &Object{
			ID: randObjectID("arr"),
			Type: &ObjectType{
				Name:       "object",
				Ref:        root.ID,
				IsRepeated: true,
			},
			Loaded: true,
		}
		root = obj
		deps = append(deps, obj)
	}
	return deps
}

// Create and register a n-dimensional(dep) array object by leaf object id
func CreateArrayObjectByID(leafObjID string, dep int, loadFn func(id string) *Object) []*Object {
	leaf := loadFn(leafObjID)
	return CreateArrayObject(leaf, dep)
}
