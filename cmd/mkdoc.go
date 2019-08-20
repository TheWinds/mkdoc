package main

import (
	"docspace"
	"docspace/scanners"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"strings"
)

var (
	scannerName = kingpin.Flag("scaner", "which api scanner to use,eg. gql-corego").Required().Short('s').String()
	tag         = kingpin.Flag("tag", "which tag to filter,eg. v1").Short('t').String()
	pkg         = kingpin.Arg("pkg", "which package to scan").Required().String()
)

var scannersMap map[string]docspace.APIScanner

func init() {
	scannerList := []docspace.APIScanner{
		new(scanners.CoregoGraphQLAPIScanner),
		new(scanners.CoregoEchoAPIScanner),
	}
	scannersMap = map[string]docspace.APIScanner{}

	for _, v := range scannerList {
		scannersMap[v.GetName()] = v
	}
}

func main() {
	kingpin.Parse()
	scanner := scannersMap[*scannerName]
	if scanner == nil {
		fmt.Printf("error: scanner \"%s\" is not found,you can choose scanner below :\n", *scannerName)
		for name := range scannersMap {
			fmt.Printf("    %s\n", name)
		}
		return
	}
	goPaths := getGOPaths()
	pkgExist := false
	for _, gopath := range goPaths {
		if _, err := os.Stat(filepath.Join(gopath, "src", *pkg)); err == nil {
			pkgExist = true
		}
	}
	if !pkgExist {
		fmt.Printf("error: package \"%s\" is not found in :\n", *pkg)
		for _, gopath := range goPaths {
			fmt.Println("  ", filepath.Join(gopath, "src", *pkg))
		}
	}
	annotations, err := scanner.ScanAnnotations(*pkg)
	if err != nil {
		fmt.Printf("error: scan annotations %v\n", err)
	}
	for _, v := range annotations {
		fmt.Println(v)
	}

}

func getGOPaths() []string {
	gopath := os.Getenv("GOPATH")
	if strings.Contains(gopath, ":") {
		return strings.Split(gopath, ":")
	}
	return []string{gopath}
}
