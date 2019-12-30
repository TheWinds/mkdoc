package mkdoc

import (
	"fmt"
	"io/ioutil"
	"os"
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
	refObjects map[string]*Object
	muObj      sync.Mutex
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
	project.refObjects = make(map[string]*Object)
	for _, obj := range BuiltinObjects() {
		project.refObjects[obj.ID] = obj
	}

	if config.BaseType != "" {
		baseTypeObj := &Object{
			ID:   config.BaseType,
			Type: &ObjectType{Name: "object"},
		}
		project.refObjects[baseTypeObj.ID] = baseTypeObj
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

func (project *Project) AddObject(id string, value *Object) {
	project.muObj.Lock()
	defer project.muObj.Unlock()
	if _, exist := project.refObjects[id]; !exist {
		project.refObjects[id] = value
	}
}

func (project *Project) GetObject(id string) *Object {
	project.muObj.Lock()
	defer project.muObj.Unlock()
	return project.refObjects[id]
}

func (project *Project) Objects() map[string]*Object {
	project.muObj.Lock()
	defer project.muObj.Unlock()
	return project.refObjects
}

func (project *Project) LoadObjects() error {
	objects := project.Objects()
	var queue []string
	for _, object := range objects {
		if !object.Loaded {
			queue = append(queue, object.ID)
		}
	}
	if len(queue) == 0 {
		return nil
	}
	i := 0
	for i < len(queue) {
		pkgType, err := newPkgType(queue[i])
		if err != nil {
			return err
		}
		err = project.loadObj(pkgType, &queue)
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

func (project *Project) getStructInfo(query *PkgType) (*GoStructInfo, error) {
	var structInfo *GoStructInfo
	var err error
	if project.Config.UseGOModule {
		pkgAbsPath := strings.Replace(query.Package, project.ModulePkg, project.ModulePath, 1)
		structInfo, err = new(StructFinder).Find(pkgAbsPath, query.TypeName)
		if err != nil {
			return nil, err
		}
		if structInfo == nil {
			return nil, fmt.Errorf("struct %s not found\n", query)
		}
		return structInfo, nil
	}

	goSrcPaths := GetGOSrcPaths()
	pkgAbsPaths := make([]string, 0)
	for _, p := range goSrcPaths {
		pkgAbsPath := filepath.Join(p, query.Package)
		pkgAbsPaths = append(pkgAbsPaths, pkgAbsPath)
		if _, err := os.Stat(pkgAbsPath); err != nil {
			continue
		}
		structInfo, err = new(StructFinder).Find(pkgAbsPath, query.TypeName)
		if err != nil && err != errGoStructNotFound {
			return nil, err
		}
		if structInfo != nil {
			break
		}
	}
	if structInfo == nil {
		return nil, fmt.Errorf("struct %s not found in any of:\n	%s", query, strings.Join(pkgAbsPaths, "\n	"))
	}
	return structInfo, nil
}

func (project *Project) loadObj(query *PkgType, queue *[]string) error {
	if query == nil {
		return nil
	}
	structInfo, err := project.getStructInfo(query)
	if err != nil {
		return err
	}

	rootObj := project.GetObject(query.fullPath)
	rootObj.Type = &ObjectType{
		Name:       "object",
		IsRepeated: false,
	}
	rootObj.Fields = make([]*ObjectField, 0)

	for _, field := range structInfo.Fields {
		if field.GoType.NotSupport{
			continue
		}
		// priority use doc comment
		var comment string
		if field.DocComment != "" {
			comment = field.DocComment
		} else {
			comment = field.Comment
		}
		fieldTag, err := NewObjectFieldTag(field.Tag)
		if err != nil {
			return err
		}
		objField := &ObjectField{
			Name: field.Name,
			Desc: comment,
			Type: &ObjectType{},
			Tag:  fieldTag,
		}
		goType := field.GoType

		// builtin type
		if goType.IsBuiltin && !goType.IsArray {
			objField.Type.Name = goType.TypeName
			rootObj.Fields = append(rootObj.Fields, objField)
			continue
		}

		objField.Type.Name = "object"

		// builtin array type
		if goType.IsBuiltin {
			arrObj := createArrayObjectByID(goType.TypeName, goType.ArrayDepth)
			objField.Type.Ref = arrObj.ID
			rootObj.Fields = append(rootObj.Fields, objField)
			continue
		}

		pkgTypePath := fmt.Sprintf("%s.%s", goType.ImportPkgName, goType.TypeName)
		obj := GetProject().GetObject(pkgTypePath)
		if obj == nil {
			obj = &Object{
				ID: pkgTypePath,
				Type: &ObjectType{
					Name:       "object",
					Ref:        "",
					IsRepeated: false,
				},
				Fields: nil,
				Loaded: false,
			}
			*queue = append(*queue, pkgTypePath)
		}

		if goType.IsArray {
			obj = createArrayObject(obj, goType.ArrayDepth)
		} else {
			obj = createArrayObject(obj, 0)
		}
		objField.Type.Ref = obj.ID

		rootObj.Fields = append(rootObj.Fields, objField)
	}
	rootObj.Loaded = true
	return nil
}
