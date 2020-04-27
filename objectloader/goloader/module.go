package goloader

import (
	"io/ioutil"
	"path/filepath"
)

type goModuleInfo struct {
	ModulePkg  string
	ModulePath string
}

func (g *GoLoader) initGoModule(pkgPath string) error {
	data, err := ioutil.ReadFile(filepath.Join(pkgPath, "go.mod"))
	if err != nil {
		return err
	}
	g.mod = &goModuleInfo{
		ModulePkg:  ModulePath(data),
		ModulePath: FindGOModAbsPath(pkgPath),
	}
	return nil
}
