package mkdoc

import (
	"fmt"
	"strings"
)

// PkgType go type and package location
type PkgType struct {
	Package  string
	TypeName string
	fullPath string
}

func (p *PkgType) parse() error {
	// github.com/a.b.c
	i := strings.LastIndex(p.fullPath, ".")
	if i == -1 {
		return fmt.Errorf("PkgType: parse '%s' error format", p.fullPath)
	}
	p.Package = p.fullPath[:i]
	p.TypeName = p.fullPath[i+1:]
	return nil
}

func newPkgType(fullPath string) (*PkgType, error) {
	pt := &PkgType{fullPath: fullPath}
	err := pt.parse()
	if err != nil {
		return nil, err
	}
	return pt, nil
}

func replacePkg(fullPath string, imports map[string]string) string {
	var i int
	for i := 0; i < len(fullPath); i += 2 {
		if fullPath[i] != '[' {
			break
		}
	}
	dot := strings.LastIndex(fullPath, ".")
	if dot == -1 {
		return fullPath[:i] + imports[""] + "." + fullPath[i:]
	}
	pkg := fullPath[i:dot]
	return fullPath[:i] + imports[pkg] + fullPath[dot:]
}
