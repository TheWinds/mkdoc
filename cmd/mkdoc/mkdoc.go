package mkdoc

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var makeDocTag *string
var makeDocVersion *string

func main() {
	app := kingpin.New("mkdoc", "make doc from go source code")

	app.
		Command("init", "").
		Action(initProject)

	cmdMake := app.Command("make", "").Action(makeDoc)
	makeDocTag = cmdMake.
		Flag("tag", "which tag to filter,eg. v1").
		Short('t').
		String()
	makeDocVersion = cmdMake.
		Flag("version", "doc version").
		Short('v').
		String()

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
