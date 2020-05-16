package main

import (
	"github.com/go-yaml/yaml"
	"io"
	"log"
	"os"
)

type config struct {
	repoName     string
	branchName   string
	gitUserName  string
	gitPassword  string
	notifyToken  string
	webUserName  string
	webPassword  string
	gopath       string
	mkdocConfigs []map[string]interface{}
	debug        bool
}

func readYamlConfig(name string) ([]map[string]interface{}, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var confs []map[string]interface{}
	dec := yaml.NewDecoder(f)
	for {
		conf := make(map[string]interface{})
		err := dec.Decode(conf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			break
		}
		confs = append(confs, conf)
	}
	return confs, nil
}

func getConfig() *config {
	conf := &config{
		gitUserName: os.Getenv("GIT_USER_NAME"),
		gitPassword: os.Getenv("GIT_PASSWORD"),
		notifyToken: os.Getenv("NOTIFY_TOKEN"),
		webUserName: os.Getenv("WEB_USER_NAME"),
		webPassword: os.Getenv("WEB_PASSWORD"),
		debug:       os.Getenv("DEBUG") == "1",
	}

	gopath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	conf.gopath = gopath

	configs, err := readYamlConfig("conf.yaml")
	if err != nil {
		log.Fatal("readConfig:", err)
	}
	if len(configs) < 2 {
		log.Fatal("miss config to mkdoc")
	}

	if name, ok := configs[0]["repo"].(string); ok {
		conf.repoName = name
	}
	if name, ok := configs[0]["branch"].(string); ok {
		conf.branchName = name
	}
	if len(conf.repoName) == 0 {
		log.Fatal("config: miss repo name")
	}
	if len(conf.branchName) == 0 {
		log.Fatal("config: miss branch name")
	}
	autoId := 0
	for _, c := range configs[1:] {
		if id, ok := c["id"].(string); !ok || len(id) == 0 {
			c["id"] = autoId
			autoId++
		}
		conf.mkdocConfigs = append(conf.mkdocConfigs, c)
	}
	return conf
}
