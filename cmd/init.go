// command: mkdoc init
package main

import (
	"docspace"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

const configFileName = "conf.yaml"

func initProject(ctx *kingpin.ParseContext) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	confFilePath := filepath.Join(path, configFileName)
	if _, err := os.Stat(confFilePath); err == nil {
		fmt.Printf("mkdoc init: config file alreadly exist")
		os.Exit(0)
	}
	cfg := &docspace.Config{
		Name:        "my doc",
		Description: "my doc project",
		APIBaseURL:  "http://",
		BodyEncoder: "json",
		CommonHeader: []*docspace.Header{
			{
				Name:    "",
				Desc:    "",
				Default: "",
			},
		},
		Package:     "",
		BaseType:    "",
		UseGOModule: false,
		Scanner:     []string{"funcdoc"},
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(confFilePath, b, 0644)
	if err != nil {
		return err
	}
	// if exist ignore
	os.Mkdir(filepath.Join(path, "docs"), 0755)
	return nil
}
