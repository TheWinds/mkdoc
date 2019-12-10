package markdown

import (
	"fmt"
	"github.com/thewinds/mkdoc"
	"strings"
)

type objGQLMarshaller struct {
	api     *mkdoc.API
	sb      *strings.Builder
	objMap  map[string]*mkdoc.Object
	rootObj *mkdoc.Object
	err     error
}

func newObjGQLMarshaller(api *mkdoc.API, obj *mkdoc.Object) *objGQLMarshaller {
	return &objGQLMarshaller{api: api, sb: new(strings.Builder), objMap: api.ObjectsMap, rootObj: obj}
}

func (o *objGQLMarshaller) Marshal() (string, error) {
	if o.err != nil {
		return "", o.err
	}
	o.marshal(o.rootObj, 0)
	return o.sb.String(), nil
}

func (o *objGQLMarshaller) marshal(obj *mkdoc.Object, dep int) {
	o.writeToken("{\n", 0)
	for _, field := range obj.Fields {
		k := fmt.Sprintf("		      %s", field.JSONTag)
		o.writeToken(k, dep)
		if !field.IsRef {
			if field.IsRepeated {
				o.writeToken("\n", 0)
			} else {
				o.writeToken("\n", 0)
			}
			continue
		}
		if field.IsRepeated {
			o.marshal(o.objMap[field.Type], dep+1)
		} else {
			o.marshal(o.objMap[field.Type], dep+1)
		}
		o.writeToken(",\n", 0)
	}
	o.writeToken("		  }", dep)
}

func (o *objGQLMarshaller) writeToken(s string, dep int) {
	prefix := ""
	for i := 0; i < dep; i++ {
		prefix += "    "
	}
	o.sb.WriteString(prefix + s)
}
