package markdown

import (
	"docspace"
	"fmt"
	"strings"
)

type objJSONMarshaller struct {
	api     *docspace.API
	sb      *strings.Builder
	objMap  map[string]*docspace.Object
	rootObj *docspace.Object
	err     error
}

func newObjJSONMarshaller(api *docspace.API, obj *docspace.Object) *objJSONMarshaller {
	return &objJSONMarshaller{api: api, sb: new(strings.Builder), objMap: api.ObjectsMap, rootObj: obj}
}

func (o *objJSONMarshaller) Marshal() (string, error) {
	if o.err != nil {
		return "", o.err
	}
	o.marshal(o.rootObj, 0)
	return o.sb.String(), nil
}

func (o *objJSONMarshaller) marshal(obj *docspace.Object, dep int) {
	if o.err != nil {
		return
	}
	if obj == nil {
		o.writeToken("null", 0)
		return
	}
	o.writeToken("{\n", 0)
	for i, field := range obj.Fields {
		k := fmt.Sprintf("    \"%s\" : ", field.JSONTag)
		o.writeToken(k, dep)
		if !field.IsRef {
			if field.IsRepeated {
				o.writeToken("[", 0)
				o.writeToken("1,2,3", 0)
				o.writeToken("]", 0)
				if i != len(obj.Fields)-1 {
					o.writeToken(",\n", 0)
				} else {
					o.writeToken("\n", 0)
				}
			} else {
				o.writeToken(docspace.MockField(field.BaseType, field.JSONTag), 0)
				if i != len(obj.Fields)-1 {
					o.writeToken(",", 0)
				}
				o.writeToken("\t# "+strings.TrimSuffix(field.Comment, "\n")+"\n", 0)
			}
			continue
		}
		if field.IsRepeated {
			o.writeToken("[", 0)
			if obj.ID != field.Type {
				o.marshal(o.objMap[field.Type], dep+1)
				o.writeToken("]", dep)
			} else {
				o.writeToken("]", 0)
			}
		} else {
			if obj.ID != field.Type {
				o.marshal(o.objMap[field.Type], dep+1)
			} else {
				o.writeToken("null", 0)
			}
		}
		if i != len(obj.Fields)-1 {
			o.writeToken(",\n", 0)
		} else {
			o.writeToken("\n", 0)
		}
	}
	o.writeToken("}", dep)
}

func (o *objJSONMarshaller) writeToken(s string, dep int) {
	prefix := ""
	for i := 0; i < dep; i++ {
		prefix += "    "
	}
	o.sb.WriteString(prefix + s)
}
