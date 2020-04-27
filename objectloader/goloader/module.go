package goloader

import (
	"io/ioutil"
	"path/filepath"
)

func (project *Project) initGoModule() error {
	data, err := ioutil.ReadFile(filepath.Join(project.Config.Package, "go.mod"))
	if err != nil {
		return err
	}
	project.ModulePkg = goloader.ModulePath(data)
	project.ModulePath = goloader.FindGOModAbsPath(project.Config.Package)
	return nil
}
