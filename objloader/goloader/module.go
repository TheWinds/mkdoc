package goloader

import (
	"github.com/thewinds/mkdoc"
	"io/ioutil"
	"path/filepath"
)

func (g *GoLoader) initGoModule(pkgPath string) error {
	data, err := ioutil.ReadFile(filepath.Join(pkgPath, "go.mod"))
	if err != nil {
		return err
	}
	g.mod = &mkdoc.GoModuleInfo{
		ModulePkg:  mkdoc.ModulePath(data),
		ModulePath: mkdoc.FindGOModAbsPath(pkgPath),
	}
	return nil
}
