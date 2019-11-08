package docspace

import (
	"fmt"
	"strings"
)

// TypeLocation go type and package location
// if this type compose to a slice or array IsRepeated should set to true
type TypeLocation struct {
	PackageName string
	TypeName    string
	IsRepeated  bool
}

func (t *TypeLocation) String() string {
	rep := ""
	if t.IsRepeated {
		rep = "[]"
	}
	return fmt.Sprintf("%s%s.%s", rep, t.PackageName, t.TypeName)
}

func newTypeLocation(raw string) *TypeLocation {
	t := &TypeLocation{}
	if strings.HasPrefix(raw, "*") {
		raw = raw[1:]
	} else if strings.HasPrefix(raw, "[]*") {
		raw = raw[3:]
		t.IsRepeated = true
	} else if strings.HasPrefix(raw, "[]") {
		t.IsRepeated = true
		raw = raw[2:]
	}
	e := strings.Split(raw, ".")
	t.PackageName, t.TypeName = e[0], e[1]
	return t
}