package mkdoc

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type mocker func(filedName string) string

var mockers map[string]mocker

/*
	"string":      true,
	"bool":        true,
	"byte":        true,
	"int":         true,
	"int32":       true,
	"int64":       true,
	"uint":        true,
	"uint32":      true,
	"uint64":      true,
	"float":       true,
	"float32":     true,
	"float64":     true,
	"interface{}": true,
*/
func init() {
	initMockers()
}
func initMockers() {
	mockers = map[string]mocker{}
	mockers["string"] = func(filedName string) string {
		if filedContainsID(filedName) {
			return fmt.Sprintf("\"%d\"", rand.Intn(999999))
		}
		if filedContainsName(filedName) {
			return "\"str\""
		}
		if filedContainsTime(filedName) {
			return time.Now().String()
		}
		return "\"\""
	}

	mockers["int"] = func(filedName string) string {
		if filedContainsID(filedName) {
			return fmt.Sprintf("%d", rand.Intn(999999))
		}
		if filedContainsTime(filedName) {
			return fmt.Sprintf("%d", time.Now().Unix())
		}
		if filedContainsPageSize(filedName) {
			return "10"
		}
		if filedContainsPage(filedName) {
			return "1"
		}
		return "0"
	}

	mockers["float"] = func(filedName string) string {
		return "0.00"
	}

	mockers["bool"] = func(filedName string) string {
		return "true"
	}

}

// MockField will mock a value according fieldType and fieldName
func MockField(fieldType, fieldName string) string {
	switch fieldType {
	case "string":
		return mockers["string"](fieldName)
	case "bool":
		return mockers["bool"](fieldName)
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return mockers["int"](fieldName)
	case "float", "float32", "float64":
		return mockers["float"](fieldName)
	case "interface{}":
		return "{}"
	default:
		return ""
	}
	return ""
}

func filedContainsID(filedName string) bool {
	idWords := []string{"id", "Id", "ID"}
	return stringContainsList(filedName, idWords)
}

func filedContainsName(filedName string) bool {
	nameWords := []string{"name", "Name"}
	return stringContainsList(filedName, nameWords)
}

func filedContainsTime(filedName string) bool {
	nameWords := []string{"time", "Time"}
	return stringContainsList(filedName, nameWords)
}

func filedContainsPageSize(filedName string) bool {
	nameWords := []string{"pageSize", "PageSize"}
	return stringContainsList(filedName, nameWords)
}

func filedContainsPage(filedName string) bool {
	nameWords := []string{"page", "current"}
	return stringContainsList(filedName, nameWords)
}

func stringContainsList(s string, words []string) bool {
	for _, v := range words {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}
