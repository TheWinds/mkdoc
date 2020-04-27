package mkdoc

import (
	"fmt"
	"github.com/thewinds/mkdoc/schema"
	"golang.org/x/sync/errgroup"
	"strings"
	"sync"
)

type Project struct {
	Config     *Config
	Scanners   []DocScanner   `yaml:"-"`
	Generators []DocGenerator `yaml:"-"`
	refObjects map[LangObjectId]*Object
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
	project.refObjects = make(map[LangObjectId]*Object)

	//if config.BaseType != "" {
	//	baseTypeObj := &Object{
	//		ID:   config.BaseType,
	//		Type: &ObjectType{Name: "object"},
	//	}
	//	project.refObjects[baseTypeObj.ID] = baseTypeObj
	//}
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
	var okScanners []DocScanner
	scanners := GetDocScanners()
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

func (project *Project) AddLangObject(id LangObjectId, value *Object) {
	project.muObj.Lock()
	defer project.muObj.Unlock()
	if _, exist := project.refObjects[id]; !exist {
		project.refObjects[id] = value
	}
}

func (project *Project) GetLangObject(id LangObjectId) *Object {
	project.muObj.Lock()
	defer project.muObj.Unlock()
	return project.refObjects[id]
}

func (project *Project) Objects() map[LangObjectId]*Object {
	project.muObj.Lock()
	defer project.muObj.Unlock()
	return project.refObjects
}

func (project *Project) parseSchemaExtension(ext *schema.Extension) (Extension, error) {
	switch ext.Name {
	case "go_tag":
		return new(ExtensionGoTag).Parse(ext)
	default:
		return new(ExtensionUnknown).Parse(ext)
	}
}

func (project *Project) parseSchemaObject(object *schema.Object) (*Object, error) {
	obj := Object{
		ID:         object.ID,
		Type:       (*ObjectType)(object.Type),
		Fields:     nil,
		Extensions: nil,
		Loaded:     false,
	}
	for _, field := range object.Fields {
		objField := &ObjectField{
			Name: field.Name,
			Desc: field.Desc,
			Type: (*ObjectType)(field.Type),
		}
		for _, ext := range field.Extensions {
			if ext != nil {
				extParsed, err := project.parseSchemaExtension(ext)
				if err != nil {
					return nil, err
				}
				objField.Extensions = append(objField.Extensions, extParsed)
			}
		}
		obj.Fields = append(obj.Fields, objField)
	}
	for _, ext := range object.Extensions {
		extParsed, err := project.parseSchemaExtension(ext)
		if err != nil {
			return nil, err
		}
		obj.Extensions = append(obj.Extensions, extParsed)
	}
	return &obj, nil
}

func (project *Project) ParseSchemaAPI(api *schema.API) (*API, error) {
	a := &API{
		API:  *api,
		Mime: &MimeType{In: api.MimeIn, Out: api.MimeOut},
	}
	if len(api.InType) > 0 {
		id := LangObjectId{Lang: api.Language, Id: api.InType}
		a.InArgument = project.GetLangObject(id)
	}
	if len(api.OutType) > 0 {
		id := LangObjectId{Lang: api.Language, Id: api.OutType}
		a.OutArgument = project.GetLangObject(id)
	}
	return a, nil
}

func (project *Project) LoadObjects(schemaDef *schema.Schema) error {
	// load object from schema object define
	for _, object := range schemaDef.Objects {
		id := LangObjectId{Lang: object.Language, Id: object.ID}
		obj, err := project.parseSchemaObject(object)
		if err != nil {
			return err
		}
		project.AddLangObject(id, obj)
	}
	langTs := make(map[string][]TypeScope)
	for _, api := range schemaDef.APIs {
		if len(api.InType) > 0 {
			ts := TypeScope{FileName: api.SourceFileName, TypeName: api.InType}
			langTs[api.Language] = append(langTs[api.Language], ts)
		}
		if len(api.OutType) > 0 {
			ts := TypeScope{FileName: api.SourceFileName, TypeName: api.OutType}
			langTs[api.Language] = append(langTs[api.Language], ts)
		}
	}
	for lang := range langTs {
		if GetObjectLoader(lang) == nil {
			return fmt.Errorf("object loader for language %s not found", lang)
		}
	}
	eg := errgroup.Group{}
	for lang, typeScopes := range langTs {
		loader := GetObjectLoader(lang)
		eg.Go(func() error {
			objs, err := loader.LoadAll(typeScopes)
			if err != nil {
				return err
			}
			for _, obj := range objs {
				id := LangObjectId{Lang: lang, Id: obj.ID}
				project.AddLangObject(id, obj)
			}
			return nil
		})
	}
	return eg.Wait()
}
