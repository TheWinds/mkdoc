package goloader

import (
	"encoding/json"
	"fmt"
	"github.com/thewinds/mkdoc"
	"testing"
)

func TestGoLoader_LoadAll(t *testing.T) {
	loader := new(GoLoader)
	loader.SetConfig(mkdoc.ObjectLoaderConfig{
		Config: mkdoc.Config{
			Path:        "",
			Name:        "",
			Description: "",
			APIBaseURL:  "",
			Injects:     nil,
			BaseType:    "",
			Scanner:     nil,
			Generator:   nil,
			Mime:        nil,
			Args:        nil,
		},
	})
	tss := []mkdoc.TypeScope{
		{
			FileName: "/Users/thewinds/develop/zhidduoke/project/src/corego/service/sale/api/class_adviser.pb.go",
			TypeName: "CreateClassAdviserReq",
		},
	}
	objects, err := loader.LoadAll(tss)
	if err != nil {
		t.Error(err)
		return
	}
	b, _ := json.MarshalIndent(objects, "", "\t")
	fmt.Println(string(b))
}
