package objmock

import (
	"fmt"
	"github.com/thewinds/mkdoc"
	"strings"
)

type JSONMocker struct {
	refs      map[string]*mkdoc.Object
	err       error
	line      int
	dep       int
	data      strings.Builder
	comment   map[int]string
	commented map[string]bool
	noRef     bool
}

func (j *JSONMocker) Mock(object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error) {
	j.refs = refs
	j.comment = make(map[int]string)
	j.commented = make(map[string]bool)
	j.dep = -1
	j.mock(object)
	if j.err != nil {
		return "", j.err
	}
	return j.appendComment(), nil
}

func (j *JSONMocker) MockNoComment(object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error) {
	j.refs = refs
	j.comment = make(map[int]string)
	j.commented = make(map[string]bool)
	j.dep = -1
	j.mock(object)
	if j.err != nil {
		return "", j.err
	}
	return j.data.String(), nil
}

func (j *JSONMocker) mock(obj *mkdoc.Object) {
	if j.err != nil {
		return
	}
	j.dep++
	j.write("{\n")
	j.line++
	var firstField bool
	for _, field := range obj.Fields {
		if field.JSONTag == "-" {
			continue
		}
		if !firstField {
			firstField = true
		} else {
			j.write(",\n")
			j.line++
		}
		j.writeIdent()
		j.write("    \"%s\": ", field.JSONTag)
		if field.IsRepeated {
			j.write("[")
		}
		j.lineComment(obj, field)
		if field.IsRef {
			refObj := j.refs[field.Type]
			if refObj == nil {
				j.err = fmt.Errorf("mock: type %s not exist", field.Type)
				return
			}
			// avoid circle ref
			if refObj.ID == field.Type {
				if j.noRef {
					j.write("null")
				} else {
					j.noRef = true
					j.mock(refObj)
					j.noRef = false
				}
			} else {
				j.mock(refObj)
			}
		} else {
			j.writeValue(field)
		}
		if field.IsRepeated {
			j.write("]")
		}
	}
	j.write("\n")
	j.line++
	j.writeIdent()
	j.write("}")
	j.dep--
}

func (j *JSONMocker) appendComment() string {
	lines := strings.Split(j.data.String(), "\n")
	var maxLen int
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	for i := range lines {
		comment := j.comment[i]
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

func (j *JSONMocker) writeValue(field *mkdoc.ObjectField) {
	val := ""
	switch field.BaseType {
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

func (j *JSONMocker) writeIdent() {
	j.data.WriteString(strings.Repeat(" ", 4*j.dep))
}

func (j *JSONMocker) lineComment(obj *mkdoc.Object, field *mkdoc.ObjectField) {
	var key string
	if field.IsRef {
		key = fmt.Sprintf("%s.%s", field.BaseType, field.Name)
	} else {
		key = fmt.Sprintf("%s.%s", obj.ID, field.Name)
	}
	if !j.commented[key] {
		j.commented[key] = true
		j.comment[j.line] = field.Comment
	}
}
