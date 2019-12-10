package mkdoc

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

type Project struct {
	Config     *Config
	Scanners   []APIScanner   `yaml:"-"`
	Generators []DocGenerator `yaml:"-"`
	ModulePkg  string
	ModulePath string
}

func NewProject(config *Config) (*Project, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	project := &Project{Config: config}
	if err := project.checkScanner(); err != nil {
		return nil, err
	}

	if err := project.checkGenerator(); err != nil {
		return nil, err
	}
	if config.UseGOModule {
		if err := project.initGoModule(); err != nil {
			return nil, err
		}
	}
	return project, nil
}

var projectOnce sync.Once
var _project *Project

func SetProject(project *Project) {
	projectOnce.Do(func() {
		_project = project
	})
}

func GetProject() *Project {
	if _project == nil {
		panic("_project is nil")
	}
	return _project
}

func (project *Project) checkScanner() error {
	var okScanners []APIScanner
	scanners := GetScanners()
	if len(project.Config.Scanner) == 0 {
		return fmt.Errorf("please use at least one scanner")
	}

	for _, name := range project.Config.Scanner {
		if scanners[name] == nil {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("scanner \"%s\" is not found,you can choose scanner below :\n", name))
			for n := range scanners {
				sb.WriteString(fmt.Sprintf("    %s\n", n))
			}
			return fmt.Errorf(sb.String())
		}
		okScanners = append(okScanners, scanners[name])
	}
	project.Scanners = okScanners
	return nil
}

func (project *Project) checkGenerator() error {
	var okGenerators []DocGenerator
	generators := GetGenerators()
	if len(project.Config.Generator) == 0 {
		return fmt.Errorf("please use at least one generator")
	}

	for _, name := range project.Config.Generator {
		if generators[name] == nil {
			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("generator \"%s\" is not found,you can choose generator below :\n", name))
			for n := range generators {
				sb.WriteString(fmt.Sprintf("    %s\n", n))
			}
			return fmt.Errorf(sb.String())
		}
		okGenerators = append(okGenerators, generators[name])
	}
	project.Generators = okGenerators
	return nil
}

func (project *Project) initGoModule() error {
	data, err := ioutil.ReadFile(filepath.Join(project.Config.Package, "go.mod"))
	if err != nil {
		return err
	}
	project.ModulePkg = ModulePath(data)
	project.ModulePath = FindGOModAbsPath(project.Config.Package)
	return nil
}
