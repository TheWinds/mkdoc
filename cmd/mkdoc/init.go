// command: mkdoc init
package mkdoc

import (
	"github.com/thewinds/mkdoc"
	"gopkg.in/alecthomas/kingpin.v2"
)

func initProject(ctx *kingpin.ParseContext) error {
	err := mkdoc.CreateDefaultConfig()
	if err != nil {
		return showWarn("init: %v", err)
	}
	return nil
}
