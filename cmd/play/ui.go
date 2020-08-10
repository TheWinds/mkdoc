package main

import (
	"log"
)

func showErr(format string, a ...interface{}) error {
	format = "❌  " + format
	log.Printf(format, a...)
	return nil
}

func showWarn(format string, a ...interface{}) error {
	format = "⚠️  " + format
	log.Printf(format, a...)
	return nil
}
