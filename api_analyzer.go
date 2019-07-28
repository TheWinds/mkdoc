package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FieldType uint32

const (
	String FieldType = 1 << iota

	Bool

	Int
	Int8
	Int16
	Int32
	Int64

	UInt
	UInt8
	UInt16
	UInt32
	UInt64

	Float
	Float32
	Float64

	Array
)

type Field struct {
	Type     FieldType
	Name     string
	Comment  string
	Required bool
}

type API struct {
	InArgument  *Field
	OutArgument *Field
}

type APIAnalyzer interface {
	GetAPIList(pkg string) ([]*API, error)
}

type CoregoAPIAnalyzer struct {
}

func (c *CoregoAPIAnalyzer) GetAPIList(pkg string) ([]*API, error) {
	goFileList, err := c.getGoFiles(pkg)
	if err != nil {
		return nil, err
	}

	for _, fileName := range goFileList {
		println(fileName)
	}
	return nil, nil
}

func (*CoregoAPIAnalyzer) getGoFiles(pkg string) ([]string, error) {
	goPath := os.Getenv("GOPATH")
	pkgPath := filepath.Join(goPath, "src", pkg)

	goFileList := make([]string, 0)

	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			if strings.HasSuffix(info.Name(), ".go") {
				goFileList = append(goFileList, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return goFileList, nil
}

func (*CoregoAPIAnalyzer) getAPICodes(fileName string) ([]string, error) {
	src, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	
}
