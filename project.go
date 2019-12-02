package docspace

type Project struct {
	Config     *Config
	Scanners   []APIScanner   `yaml:"-"`
	Generators []DocGenerator `yaml:"-"`
}

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
