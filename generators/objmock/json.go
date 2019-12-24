package objmock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/thewinds/mkdoc"
	"strings"
)

type JSONMocker struct {
	refs      map[string]*mkdoc.Object
	err       error
	dep       int
	data      strings.Builder
	comment   map[int]string
	commented map[string]bool
	fieldNo   int
	refPath   []string
}

func NewJSONMocker() *JSONMocker {
	return &JSONMocker{}
}

func (j *JSONMocker) Mock(object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error) {
	j.refs = refs
	j.comment = make(map[int]string)
	j.commented = make(map[string]bool)
	j.dep = -1
	j.pushRefPath(object.ID)
	j.mock(object)
	if j.err != nil {
		return "", j.err
	}
	return j.data.String(), nil
}

func (j *JSONMocker) MockPretty(object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error) {
	r, err := j.Mock(object, refs)
	if err != nil {
		return "", err
	}
	dst := bytes.NewBuffer([]byte{})
	err = json.Indent(dst, []byte(r), "", "    ")
	if err != nil {
		return "", err
	}
	return dst.String(), nil
}

func (j *JSONMocker) MockPrettyComment(object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error) {
	r, err := j.MockPretty(object, refs)
	if err != nil {
		return "", err
	}
	return j.appendComment(r), nil
}

func (j *JSONMocker) mock(obj *mkdoc.Object) {
	if j.err != nil {
		return
	}
	objType := obj.Type
	if objType.IsRepeated {
		j.write("[")
		defer func() { j.write("]") }()
	}

	// builtin type
	if objType.Name != "object" {
		j.writeValue(objType.Name)
		return
	}

	if objType.Ref != "" {
		j.mockRef(objType.Ref)
		return
	}
	j.write("{")
	defer func() { j.write("}") }()
	var firstField bool
	for _, field := range obj.Fields {
		jsonTag := field.Tag.GetTagName("json")
		if jsonTag == "-" {
			continue
		}
		if jsonTag == "" {
			jsonTag = field.Name
		}
		if !firstField {
			firstField = true
		} else {
			j.write(",")
		}
		j.write("\"%s\":", jsonTag)

		fieldTyp := field.Type
		j.markFieldComment(obj, field)
		j.fieldNo++
		if fieldTyp.Ref != "" {
			j.mockRef(fieldTyp.Ref)
		} else {
			j.writeValue(fieldTyp.Name)
		}
	}
}

func (j *JSONMocker) mockRef(refID string) {
	refObj := j.refs[refID]
	if refObj == nil {
		j.err = fmt.Errorf("mock: type %s not exist", refID)
		return
	}
	// limit circle ref
	if j.pushRefPath(refObj.ID) {
		j.mock(refObj)
		j.popRefPath()
	} else {
		j.write("null")
	}
}

func (j *JSONMocker) appendComment(src string) string {
	lines := strings.Split(src, "\n")
	var maxLen int
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	fieldNo := -1
	for i := range lines {
		if strings.Index(lines[i], "\":") == -1 {
			continue
		}
		fieldNo++
		comment := j.comment[fieldNo]
		if comment != "" {
			lines[i] = fmt.Sprintf("%s // %s", j.padRight(lines[i], maxLen), comment)
		}
	}
	return strings.Join(lines, "\n")
}

func (j *JSONMocker) padRight(s string, l int) string {
	p := l - len(s)
	return s + strings.Repeat(" ", p)
}

func (j *JSONMocker) write(format string, i ...interface{}) {
	fmt.Fprintf(&j.data, format, i...)
}

func (j *JSONMocker) writeValue(typ string) {
	val := ""
	switch typ {
	case "string":
		val = "\"str\""
	case "bool":
		val = "true"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		val = "10"
	case "float", "float32", "float64":
		val = "10.2"
	case "interface{}":
		val = "{}"
	default:
		val = ""
	}
	j.write(val)
}

func (j *JSONMocker) markFieldComment(obj *mkdoc.Object, field *mkdoc.ObjectField) {
	var key string
	key = fmt.Sprintf("%s.%s", obj.ID, field.Name)
	if !j.commented[key] {
		j.commented[key] = true
		j.comment[j.fieldNo] = field.Desc
	}
}

func (j *JSONMocker) pushRefPath(objTyp string) bool {
	i := -1
	for k, v := range j.refPath {
		if v == objTyp {
			i = k
			break
		}
	}
	// a->a        ok
	// a->a->a 	  !ok
	// a->a->b->a !ok
	if i == -1 || len(j.refPath)-i <= 1 {
		j.refPath = append(j.refPath, objTyp)
		return true
	}
	return false
}

func (j *JSONMocker) popRefPath() {
	if len(j.refPath) == 0 {
		return
	}
	j.refPath = j.refPath[:len(j.refPath)-1]
}
