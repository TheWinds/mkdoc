package objmock

import (
	"fmt"
	"strings"

	"github.com/thewinds/mkdoc"
)

type GQLBodyMocker struct {
	refs    map[string]*mkdoc.Object
	err     error
	dep     int
	data    strings.Builder
	refPath []string
}

func GqlBodyMocker() *GQLBodyMocker {
	return &GQLBodyMocker{}
}

func (g *GQLBodyMocker) Mock(object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error) {
	if object == nil {
		return "\n", nil
	}
	g.refs = refs
	g.dep = -1
	g.pushRefPath(object.ID)
	g.mock(object)
	if g.err != nil {
		return "", g.err
	}
	return g.data.String(), nil
}

func (g *GQLBodyMocker) MockPretty(object *mkdoc.Object, refs map[string]*mkdoc.Object, prefix string, indent string) (string, error) {
	if object == nil {
		return "\n", nil
	}
	r, err := g.Mock(object, refs)
	if err != nil {
		return "", err
	}
	return g.indent(r, prefix, indent), nil
}

func (g *GQLBodyMocker) indent(s string, prefix string, indent string) string {
	sb := strings.Builder{}
	dep := 0
	var w []rune
	for _, c := range s {
		switch c {
		case '{':
			sb.WriteString(prefix)
			sb.WriteString(strings.Repeat(indent, dep))
			if len(w) > 0 {
				sb.WriteString(string(w))
				w = nil
			}
			dep++
			sb.WriteRune(c)
			sb.WriteRune('\n')
			//sb.WriteString(strings.Repeat(indent, dep))
		case '}':
			dep--
			sb.WriteString(prefix)
			sb.WriteString(strings.Repeat(indent, dep))
			sb.WriteRune(c)
			sb.WriteRune('\n')
		case '\n':
			sb.WriteString(prefix)
			sb.WriteString(strings.Repeat(indent, dep))
			sb.WriteString(string(w))
			w = nil
			sb.WriteRune('\n')
		default:
			w = append(w, c)
		}
	}
	return sb.String()
}

func (g *GQLBodyMocker) mock(obj *mkdoc.Object) {
	if g.err != nil {
		return
	}
	objType := obj.Type
	if objType.IsRepeated {
		//g.write("{")
		//defer func() { g.write("}") }()
	}

	// builtin type
	if objType.Name != "object" {
		g.writeValue()
		return
	}

	if objType.Ref != "" {
		g.mockRef(objType.Ref)
		return
	}
	g.write("{")
	defer func() { g.write("}") }()
	for _, field := range obj.Fields {
		jsonTag := field.Tag.GetFirstValue("json", ",")
		if jsonTag == "-" {
			continue
		}
		if jsonTag == "" {
			jsonTag = field.Name
		}
		g.write(jsonTag)

		fieldTyp := field.Type
		if fieldTyp.Ref != "" {
			g.mockRef(fieldTyp.Ref)
		} else {
			g.writeValue()
		}
	}
}

func (g *GQLBodyMocker) mockRef(refID string) {
	refObj := g.refs[refID]
	if refObj == nil {
		g.err = fmt.Errorf("mock: type %s not exist", refID)
		return
	}
	// limit circle ref
	if g.pushRefPath(refObj.ID) {
		g.mock(refObj)
		g.popRefPath()
	} else {
		g.write("{}")
	}
}

func (g *GQLBodyMocker) write(format string, i ...interface{}) {
	fmt.Fprintf(&g.data, format, i...)
}

func (g *GQLBodyMocker) writeValue() {
	g.write("\n")
}

func (g *GQLBodyMocker) pushRefPath(objTyp string) bool {
	i := -1
	for k, v := range g.refPath {
		if v == objTyp {
			i = k
			break
		}
	}
	// a->a        ok
	// a->a->a 	  !ok
	// a->a->b->a !ok
	if i == -1 || len(g.refPath)-i <= 1 {
		g.refPath = append(g.refPath, objTyp)
		return true
	}
	return false
}

func (g *GQLBodyMocker) popRefPath() {
	if len(g.refPath) == 0 {
		return
	}
	g.refPath = g.refPath[:len(g.refPath)-1]
}
