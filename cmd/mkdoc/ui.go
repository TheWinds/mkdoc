package main

import "fmt"

func showErr(format string, a ...interface{}) error {
	format = "❌  " + format
	_, err := fmt.Printf(format, a...)
	return err
}

func showWarn(format string, a ...interface{}) error {
	format = "⚠️  " + format
	_, err := fmt.Printf(format, a...)
	return err
}
