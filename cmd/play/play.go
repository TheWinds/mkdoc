// +build js

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall/js"
)

func main() {
	initJS()
	loadSource()
	err := makeDoc()
	if err != nil {
		log.Println("mkdoc:", err)
		return
	}
	setGenResult()
}

type ConsoleWriter struct {
}

func (c *ConsoleWriter) Write(p []byte) (n int, err error) {
	js.Global().Get("console").Call("log", string(p))
	return len(p), nil
}

func (c *ConsoleWriter) Log(s string) {
	c.Write([]byte(s))
}

var console = new(ConsoleWriter)

func initJS() {
	log.SetOutput(console)
	os.Setenv("PWD", "/")
}

func loadSource() {
	dom := js.Global().Get("document")
	defaultConf := unQuote(dom.Call("getElementById", "code-conf").Get("value").String())
	apiGO := unQuote(dom.Call("getElementById", "code-api").Get("value").String())
	userGO := unQuote(dom.Call("getElementById", "code-user").Get("value").String())
	const modSrc = `module github.com/thewinds/mkdoc/example

go 1.12
`

	ioutil.WriteFile("conf.yaml", []byte(defaultConf), 0666)
	os.MkdirAll("src/model", 0755)
	ioutil.WriteFile("src/model/user.go", []byte(userGO), 0666)
	os.MkdirAll("src/view", 0755)
	ioutil.WriteFile("src/api.go", []byte(apiGO), 0666)
	ioutil.WriteFile("src/go.mod", []byte(modSrc), 0666)
}

func unQuote(s string) string {
	if strings.HasPrefix(s, `""`) && strings.HasSuffix(s, `""`) && len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return s
}

func setGenResult() {
	var mds []string
	filepath.Walk("./docs/docsify/", func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && strings.HasSuffix(path, ".md") {
			mds = append(mds, path)
		}
		return err
	})
	m := make(map[string]interface{})
	for _, v := range mds {
		data, _ := ioutil.ReadFile(v)
		m["/"+filepath.Base(v)] = string(data)
		err := os.Remove(v)
		if err != nil {
			log.Println("rm:", v)
		}
	}
	data, _ := json.Marshal(m)
	js.Global().Call("reloadDoc", string(data))
}
