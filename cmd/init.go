// command: mkdoc init
package main

import (
	"docspace"
	"gopkg.in/alecthomas/kingpin.v2"
)

func initProject(ctx *kingpin.ParseContext) error {
	err := docspace.CreateDefaultConfig()
	if err != nil {
		return showWarn("init: %v", err)
	}
	return nil
}
