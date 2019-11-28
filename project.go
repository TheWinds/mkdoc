package docspace

type Project struct {
	Name         string    `yaml:"name"`
	Description  string    `yaml:"desc"`
	APIBaseURL   string    `yaml:"api_base_url"`  // https://api.xxx.com
	BodyEncoder  string    `yaml:"body_encoder"`  // json,xml,form
	CommonHeader []*Header `yaml:"common_header"` //
	BasePackage  string    `yaml:"pkg"`           //
	BaseType     string    `yaml:"base_type"`     // models.BaseType
	UseGOModule  bool      `yaml:"use_go_mod"`
	Scanner      []string  `yaml:"scanner"`

	OnScanners []APIScanner `yaml:"-"`
}

type Header struct {
	Name    string `yaml:"name"`
	Desc    string `yaml:"desc"`
	Default string `yaml:"default"`
}
