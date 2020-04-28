package gqlboss

import (
	"docspace"
	"fmt"
	"testing"
)

func TestScanner_ScanAnnotations(t *testing.T) {
	scanner := new(Scanner)
	a, err := scanner.ScanAnnotations(docspace.Project{Config: &{Package: "corego"}})
	if err != nil {
		t.Error(err)
		return
	}
	for _, v := range a {
		fmt.Println(v)
	}
}
