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
		if field.Tag != nil {
			newField.Tag = mustObjectFieldTag(field.Tag.raw)
		}
		newObj.Fields = append(newObj.Fields, newField)
	}
	newObj.Loaded = obj.Loaded
	return newObj
}

func randObjectID(s string) string {
	return fmt.Sprintf("obj_%s_#%d", s, rand.Int63())
}

// createRootObject
// crate root object by package and type
// at the same register those created object to project's ref objects
func createRootObject(pkgTyp string) (*Object, error) {
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

	leaf := GetProject().GetObject(pkgTypPath)
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
	return createArrayObject(leaf, arrDep), nil
}

// create and register a n-dimensional(dep) array object
func createArrayObject(leaf *Object, dep int) *Object {
	root := leaf
	GetProject().AddObject(root.ID, root)
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
		GetProject().AddObject(root.ID, root)
	}
	return root
}

// create and register a n-dimensional(dep) array object by leaf object id
func createArrayObjectByID(leafObjID string, dep int) *Object {
	leaf := GetProject().GetObject(leafObjID)
	return createArrayObject(leaf, dep)
}
