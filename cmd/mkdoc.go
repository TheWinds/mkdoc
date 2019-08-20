package main

import (
	"docspace"
	"docspace/scanners"
	"gopkg.in/alecthomas/kingpin.v2"
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
	for _, v := range scannerList {
		scannersMap[v.GetName()] = v
	}
}

func main() {
	kingpin.Parse()
	if scannersMap[*scannerName] == nil {

	}
}
