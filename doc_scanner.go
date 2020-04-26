package mkdoc

import (
	"github.com/thewinds/mkdoc/schema"
	"log"
)

type DocScanner interface {
	Scan(config Config) (*DocScanResult, error)
	Name() string
	Help() string
}

type DocScanResult struct {
	APIs    []*schema.API
	Objects map[LangObjectId]*schema.Object
}

type DocScanConfig struct {
	ProjectConfig Config
	Args          map[string]string
}

var docScanners map[string]DocScanner

// RegisterDocScanner to global doc scanners
func RegisterDocScanner(scanner DocScanner) {
	if docScanners == nil {
		docScanners = make(map[string]DocScanner)
	}
	scannerName := scanner.Name()
	if docScanners[scannerName] != nil {
		log.Fatalf("duplicate register doc scanner : %s", scannerName)
	}
	docScanners[scannerName] = scanner
}

// GetScanners get all registered scanners
func GetDocScanners() map[string]DocScanner {
	return docScanners
}
