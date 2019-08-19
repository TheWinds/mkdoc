package main

import (
	"fmt"
	"regexp"
)

func main() {
	var re = regexp.MustCompile(`(\w+)\s+(\w+)\s*(.+)*`)
	var str = `name string 名称`
	for i, match := range re.FindStringSubmatch(str) {
		fmt.Println(match, "found at index", i)
	}
	for _, v := range re.SubexpNames() {
		fmt.Println(v)
	}
	return
	scanGraphQLAPIDocInfo("corego/service/boss/schemas")
}