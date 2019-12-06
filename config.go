package docspace

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const configFileName = "conf.yaml"

type Header struct {
	Name    string `yaml:"name"`
	Desc    string `yaml:"desc"`
	Default string `yaml:"default"`
}

type Config struct {
	Name         string    `yaml:"name"`
	Description  string    `yaml:"desc"`
	APIBaseURL   string    `yaml:"api_base_url"`  // https://api.xxx.com
	BodyEncoder  string    `yaml:"body_encoder"`  // json,xml,form
	CommonHeader []*Header `yaml:"common_header"` //
	Package      string    `yaml:"pkg"`           //
	BaseType     string    `yaml:"base_type"`     // models.BaseType
	UseGOModule  bool      `yaml:"use_go_mod"`
	Scanner      []string  `yaml:"scanner"`
	Generator    []string  `yaml:"generator"`
}

func (config *Config) Check() error {
	// check if the pkg to scan is exist
	if config.Package == "" {
		return fmt.Errorf("please config a pkg to scan in conf.yaml")
	}

	if config.UseGOModule {
		path := config.Package
		if !filepath.IsAbs(path) {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			path = filepath.Join(wd, path)
		}
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("no such file or directory: %s\n", path)
		}
	} else {
		goPaths := GetGOPaths()
		pkgExist := false
		for _, gopath := range goPaths {
			if _, err := os.Stat(filepath.Join(gopath, "src", config.Package)); err == nil {
				pkgExist = true
			}
		}
		if !pkgExist {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("error: package \"%s\" is not found in any of:\n", config.Package))
			for _, gopath := range goPaths {
				sb.WriteString(fmt.Sprintln("  ", filepath.Join(gopath, "src", config.Package)))
			}
			return fmt.Errorf("%s", sb.String())
		}
	}
	return nil
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
	return config, err
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
		BodyEncoder: "json",
		CommonHeader: []*Header{
			{
				Name:    "",
				Desc:    "",
				Default: "",
			},
		},
		Package:     "",
		BaseType:    "",
		UseGOModule: false,
		Scanner:     []string{"funcdoc"},
		Generator:   []string{"markdown"},
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
