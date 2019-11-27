package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var makeDocTag *string
var makeDocOut *string

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
	makeDocOut = cmdMake.
		Flag("out", "out file name").
		Short('o').
		String()

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
