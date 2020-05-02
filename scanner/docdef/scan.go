package docdef

import (
	"encoding/json"
	"github.com/thewinds/mkdoc"
	"github.com/thewinds/mkdoc/schema"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	mkdoc.RegisterDocScanner(new(Scanner))
}

type Scanner struct {
	filterTag string
	path      string
	fileExt   string
}

func (s *Scanner) Scan(config mkdoc.DocScanConfig) (*mkdoc.DocScanResult, error) {
	s.filterTag = config.Args["_filter_tag"]
	s.path = config.Args["path"]
	s.fileExt = config.Args["file_ext"]
	if len(s.fileExt) == 0 {
		s.fileExt = ".doc.json"
	}
	r := new(mkdoc.DocScanResult)
	defs, err := s.scanDefs(&config)
	if err != nil {
		return nil, err
	}
	for _, def := range defs {
		for _, api := range def.APIs {
			ok := false
			for _, tag := range api.Tags {
				if tag == s.filterTag || len(s.filterTag) == 0 {
					ok = true
					break
				}
			}
			if ok {
				r.APIs = append(r.APIs, api)
			}
		}
		r.Objects = append(r.Objects, def.Objects...)
	}
	return r, nil
}

func (s *Scanner) scanDefs(config *mkdoc.DocScanConfig) ([]*schema.Schema, error) {
	var schemas []*schema.Schema
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, s.fileExt) {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		s := new(schema.Schema)
		err = json.Unmarshal(b, s)
		if err != nil {
			return err
		}
		if len(s.APIs) == 0 {
			return nil
		}
		schemas = append(schemas, s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return schemas, nil
}

func (s *Scanner) Name() string {
	return "docdef"
}

func (s *Scanner) Help() string {
	return "scan code from doc schema json"
}
