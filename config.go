package mkdoc

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
)

const configFileName = "conf.yaml"

type Inject struct {
	Name    string `yaml:"name"`
	Desc    string `yaml:"desc"`
	Default string `yaml:"default"`
	Scope   string `yaml:"scope"` // header,query,form
}

type MimeType struct {
	In  string `yaml:"in"`
	Out string `yaml:"out"`
}

type Config struct {
	Path        string            `yaml:"path"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"desc"`
	APIBaseURL  string            `yaml:"api_base_url"` // https://api.xxx.com
	Injects     []*Inject         `yaml:"inject"`       //
	Scanner     []string          `yaml:"scanner"`
	Generator   []string          `yaml:"generator"`
	Mime        *MimeType         `yaml:"mime"` // MimeType
	Args        map[string]string `yaml:"args"`
}

func (config *Config) Check() error {
	// check if the pkg to scan is exist
	if config.Path == "" {
		return fmt.Errorf("please config a path to scan in conf.yaml")
	}
	return nil
}

func (config *Config) Copy() Config {
	c := Config{
		Path:        config.Path,
		Name:        config.Name,
		Description: config.Description,
		APIBaseURL:  config.APIBaseURL,
		Injects:     nil,
		Scanner:     nil,
		Generator:   nil,
		Mime:        nil,
		Args:        nil,
	}
	for _, inject := range config.Injects {
		cp := *inject
		c.Injects = append(c.Injects, &cp)
	}
	for _, s := range config.Scanner {
		c.Scanner = append(c.Scanner, s)
	}
	for _, s := range config.Generator {
		c.Generator = append(c.Generator, s)
	}
	mime := *config.Mime
	c.Mime = &mime
	c.Args = make(map[string]string, len(config.Args))
	for k, v := range config.Args {
		c.Args[k] = v
	}
	return c
}

func loadConfig(fileName string) (*Config, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	confFilePath := filepath.Join(path, fileName)
	b, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}
	config := new(Config)
	err = yaml.Unmarshal(b, config)
	return mergeDefault(config), err
}

func mergeDefault(conf *Config) *Config {
	if conf == nil {
		return nil
	}
	if conf.Mime == nil {
		conf.Mime = &MimeType{"form", "json"}
	}
	return conf
}

func LoadDefaultConfig() (*Config, error) {
	return loadConfig(configFileName)
}

func CreateDefaultConfig() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	confFilePath := filepath.Join(path, configFileName)
	if _, err := os.Stat(confFilePath); err == nil {
		return fmt.Errorf("config file alreadly exist")
	}
	cfg := &Config{
		Name:        "my doc",
		Description: "this doc is auto generated by [mkdoc](https://github.com/TheWinds/docspace)",
		APIBaseURL:  "http://",
		Mime: &MimeType{
			In:  "form",
			Out: "json",
		},
		Injects: []*Inject{
			{
				Name:    "",
				Desc:    "",
				Default: "",
				Scope:   "",
			},
		},
		Path:      "",
		Scanner:   []string{"funcdoc"},
		Generator: []string{"markdown"},
		Args:      map[string]string{},
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(confFilePath, b, 0644)
	if err != nil {
		return err
	}
	// if exist ignore
	os.Mkdir(filepath.Join(path, "docs"), 0755)
	return nil
}
