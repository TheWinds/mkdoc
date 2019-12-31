package objmock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/thewinds/mkdoc"
)

type ZKGQLMocker struct {
	refs      map[string]*mkdoc.Object
	err       error
	dep       int
	data      strings.Builder
	comment   map[int]string
	commented map[string]bool
	fieldNo   int
	refPath   []string
}

func NewZKGQLMocker() *ZKGQLMocker {
	return &ZKGQLMocker{}
}

/*
query courses($current: Int!, $pageSize: Int!, $courseName: String, $isOnline: Int, $courseClass: Int) {
  courses(current: $current, pageSize: $pageSize, courseName: $courseName, isOnline: $isOnline, courseClass: $courseClass) {
    total
    bodys{
      courseId
      courseName
      courseType
      courseClass
      courseCover
      pcCourseCover
      price
      isOnline
      introduce
      qrCode
      deadlineType
      bindTeacherIds
      linkAfterPurchase
      hasPhysicalProduct
      contactWechat
      createTime
      brief
      videoType
      headCover
      courseItemPics
      courseBuyPics

    }
    errorCode
    errorMsg
    success
  }
}
*/

func (z *ZKGQLMocker) Mock(name string, object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error) {
	if object == nil {
		return "\n", nil
	}
	z.refs = refs
	z.comment = make(map[int]string)
	z.commented = make(map[string]bool)
	z.dep = -1
	z.pushRefPath(object.ID)
	z.mock(object)
	if z.err != nil {
		return "", z.err
	}
	return z.data.String(), nil
}

func (z *ZKGQLMocker) MockPretty(name string, object *mkdoc.Object, refs map[string]*mkdoc.Object) (string, error) {
	if object == nil {
		return "\n", nil
	}
	r, err := z.Mock(name, object, refs)
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

func (z *ZKGQLMocker) mockHeader(name string, obj *mkdoc.Object) {
	for _, field := range obj.Fields {
		fType := field.Type
		if fType.IsRepeated || fType.Name == "object" {
			fmt.Printf("gql_zk: SKIP '%s' field %s array or object field is not support", name, field.Name)
			continue
		}

	}
}

func (z *ZKGQLMocker) mock(obj *mkdoc.Object) {
	if z.err != nil {
		return
	}
	objType := obj.Type
	if objType.IsRepeated {
		z.write("[")
		defer func() { z.write("]") }()
	}

	// builtin type
	if objType.Name != "object" {
		z.writeValue(objType.Name)
		return
	}

	if objType.Ref != "" {
		z.mockRef(objType.Ref)
		return
	}
	z.write("{")
	defer func() { z.write("}") }()
	var firstField bool
	for _, field := range obj.Fields {
		jsonTag := field.Tag.GetFirstValue("json", ",")
		if jsonTag == "-" {
			continue
		}
		if jsonTag == "" {
			jsonTag = field.Name
		}
		if !firstField {
			firstField = true
		} else {
			z.write(",")
		}
		z.write("\"%s\":", jsonTag)

		fieldTyp := field.Type
		z.markFieldComment(obj, field)
		z.fieldNo++
		if fieldTyp.Ref != "" {
			z.mockRef(fieldTyp.Ref)
		} else {
			z.writeValue(fieldTyp.Name)
		}
	}
}

func (z *ZKGQLMocker) mockRef(refID string) {
	refObj := z.refs[refID]
	if refObj == nil {
		z.err = fmt.Errorf("mock: type %s not exist", refID)
		return
	}
	// limit circle ref
	if z.pushRefPath(refObj.ID) {
		z.mock(refObj)
		z.popRefPath()
	} else {
		z.write("null")
	}
}

func (z *ZKGQLMocker) appendComment(src string) string {
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
		comment := z.comment[fieldNo]
		if comment != "" {
			lines[i] = fmt.Sprintf("%s // %s", z.padRight(lines[i], maxLen), comment)
		}
	}
	return strings.Join(lines, "\n")
}

func (z *ZKGQLMocker) padRight(s string, l int) string {
	p := l - len(s)
	return s + strings.Repeat(" ", p)
}

func (z *ZKGQLMocker) write(format string, i ...interface{}) {
	fmt.Fprintf(&z.data, format, i...)
}

func (z *ZKGQLMocker) writeValue(typ string) {
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
	z.write(val)
}

func (z *ZKGQLMocker) markFieldComment(obj *mkdoc.Object, field *mkdoc.ObjectField) {
	var key string
	key = fmt.Sprintf("%s.%s", obj.ID, field.Name)
	if !z.commented[key] {
		z.commented[key] = true
		z.comment[z.fieldNo] = field.Desc
	}
}

func (z *ZKGQLMocker) pushRefPath(objTyp string) bool {
	i := -1
	for k, v := range z.refPath {
		if v == objTyp {
			i = k
			break
		}
	}
	// a->a        ok
	// a->a->a 	  !ok
	// a->a->b->a !ok
	if i == -1 || len(z.refPath)-i <= 1 {
		z.refPath = append(z.refPath, objTyp)
		return true
	}
	return false
}

func (z *ZKGQLMocker) popRefPath() {
	if len(z.refPath) == 0 {
		return
	}
	z.refPath = z.refPath[:len(z.refPath)-1]
}
