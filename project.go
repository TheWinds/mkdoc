package mkdoc

import (
	"fmt"
	"github.com/thewinds/mkdoc/objectloader/goloader"
	"github.com/thewinds/mkdoc/schema"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
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

func (project *Project) LoadObjects11(ids ...LangObjectId) error {
	objects := project.Objects()
	var queue []string
	if len(ids) == 0 {
		// load all
		for _, object := range objects {
			if !object.Loaded {
				queue = append(queue, object.ID)
			}
		}
	} else {
		if project.Config.BaseType != "" {
			queue = append(queue, project.Config.BaseType)
		}
		for _, id := range ids {
			toLoadID := project.firstUnLoadID(objects, id)
			if toLoadID != "" {
				queue = append(queue, toLoadID)
			}
		}
	}

	if len(queue) == 0 {
		return nil
	}
	i := 0
	for i < len(queue) {
		pkgType, err := goloader.newPkgType(queue[i])
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

func (project *Project) getStructInfo(query *goloader.PkgType) (*goloader.GoStructInfo, error) {
	var structInfo *goloader.GoStructInfo
	var err error
	if project.Config.UseGOModule {
		pkgAbsPath := strings.Replace(query.Package, project.ModulePkg, project.ModulePath, 1)
		structInfo, err = new(goloader.StructFinder).Find(pkgAbsPath, query.TypeName)
		if err != nil {
			return nil, err
		}
		if structInfo == nil {
			return nil, fmt.Errorf("struct %s not found\n", query)
		}
		return structInfo, nil
	}

	goSrcPaths := goloader.GetGOSrcPaths()
	pkgAbsPaths := make([]string, 0)
	for _, p := range goSrcPaths {
		pkgAbsPath := filepath.Join(p, query.Package)
		pkgAbsPaths = append(pkgAbsPaths, pkgAbsPath)
		if _, err := os.Stat(pkgAbsPath); err != nil {
			continue
		}
		structInfo, err = new(goloader.StructFinder).Find(pkgAbsPath, query.TypeName)
		if err != nil && err != goloader.errGoStructNotFound {
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

func (project *Project) loadObj(query *goloader.PkgType, queue *[]string) error {
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
		if field.GoType.NotSupport {
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

// dfs search
func (project *Project) firstUnLoadID(objects map[string]*Object, id string) string {
	obj := objects[id]
	if obj == nil {
		return ""
	}
	if !obj.Loaded {
		return id
	}
	if obj.Type.Ref == "" {
		return ""
	}
	return project.firstUnLoadID(objects, obj.Type.Ref)
}
