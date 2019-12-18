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

func NewPkgType(fullPath string) (*PkgType, error) {
	pt := &PkgType{fullPath: fullPath}
	err := pt.parse()
	if err != nil {
		return nil, err
	}
	return pt, nil
}
