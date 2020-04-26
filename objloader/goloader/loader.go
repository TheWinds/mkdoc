package goloader

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thewinds/mkdoc"
	"github.com/thewinds/mkdoc/schema"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func init() {
	mkdoc.RegisterObjectLoader(new(GoLoader))
}

type GoLoader struct {
	config      *mkdoc.ObjectLoaderConfig
	cached      map[string]*mkdoc.Object
	tsId        map[mkdoc.TypeScope]string
	initialed   bool
	once        sync.Once
	mod         *mkdoc.GoModuleInfo
	enableGoMod bool
}

func (g *GoLoader) init(config *mkdoc.ObjectLoaderConfig) {
	g.once.Do(func() {
		g.config = config
		g.cached = make(map[string]*mkdoc.Object)
		g.tsId = make(map[mkdoc.TypeScope]string)
		g.initialed = true
		if config.Args[EnableGoModule] == "true" {
			g.enableGoMod = true
		}
		for _, object := range BuiltinObjects() {
			g.cached[object.ID] = object
		}
	})
}

func (g *GoLoader) loadCache(id string) *mkdoc.Object {
	if !g.initialed {
		return nil
	}
	return g.cached[id]
}

func (g *GoLoader) Add(object *mkdoc.Object) error {
	if !g.initialed {
		return errors.New("loader not initialed")
	}
	g.cached[object.ID] = object
	return nil
}

func (g *GoLoader) GetObjectId(ts mkdoc.TypeScope) (string, error) {
	if !g.initialed {
		return "", errors.New("loader not initialed")
	}
	if g.enableGoMod && (g.mod == nil) {
		if err := g.initGoModule(g.config.Path); err != nil {
			return "", err
		}
	}
	if g.tsId[ts] == "" {
		imports, err := mkdoc.GetFileImportsAtFile(ts.FileName, g.mod)
		if err != nil {
			return "", err
		}
		g.tsId[ts] = replacePkg(ts.TypeName, imports)
	}
	return g.tsId[ts], nil
}

func (g *GoLoader) LoadAll(tss []mkdoc.TypeScope) ([]*mkdoc.Object, error) {
	if !g.initialed {
		return nil, errors.New("loader not initialed")
	}
	if g.enableGoMod && (g.mod == nil) {
		if err := g.initGoModule(g.config.Path); err != nil {
			return nil, err
		}
	}
	var unloads []*mkdoc.Object
	for _, ts := range tss {
		pkgTyp, err := g.GetObjectId(ts)
		if err != nil {
			return nil, err
		}
		objs, err := mkdoc.CreateRootObject(pkgTyp, g.loadCache)
		if err != nil {
			return nil, err
		}
		for _, obj := range objs {
			g.cached[obj.ID] = obj
			if !obj.Loaded {
				unloads = append(unloads, objs...)
			}
		}
	}
	if err := g.loadUnloads(unloads); err != nil {
		return nil, err
	}
	var r []*mkdoc.Object
	for _, object := range g.cached {
		r = append(r, object)
	}
	return r, nil
}

func (g *GoLoader) loadUnloads(unloads []*mkdoc.Object) error {
	if len(unloads) == 0 {
		return nil
	}
	var queue []string
	for _, obj := range unloads {
		toLoadID := g.lookupUnLoadId(obj.ID)
		if toLoadID != "" {
			queue = append(queue, toLoadID)
		}
	}
	i := 0
	for i < len(queue) {
		id := queue[i]
		pkgType, err := newPkgType(id)
		if err != nil {
			return err
		}
		err = g.loadObj(pkgType, &queue)
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

func (g *GoLoader) lookupUnLoadId(id string) string {
	o := g.loadCache(id)
	if o == nil {
		return ""
	}
	if !o.Loaded {
		return o.ID
	}
	if o.Type.Ref == "" {
		return ""
	}
	return g.lookupUnLoadId(o.Type.Ref)
}

func (g *GoLoader) loadObj(query *PkgType, queue *[]string) error {
	if query == nil {
		return nil
	}
	structInfo, err := g.getStructInfo(query)
	if err != nil {
		return err
	}

	rootObj := g.loadCache(query.fullPath)
	rootObj.Type = &mkdoc.ObjectType{
		Name:       "object",
		IsRepeated: false,
	}
	rootObj.Fields = make([]*mkdoc.ObjectField, 0)

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
		fieldTagExt, err := new(mkdoc.ExtensionGoTag).Parse(&schema.Extension{
			Name: "go_tag",
			Data: json.RawMessage(fmt.Sprintf("%q", field.Tag)),
		})
		if err != nil {
			return err
		}
		objField := &mkdoc.ObjectField{
			Name:       field.Name,
			Desc:       comment,
			Type:       &mkdoc.ObjectType{},
			Extensions: []mkdoc.Extension{fieldTagExt},
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
			arrObjs := mkdoc.CreateArrayObjectByID(goType.TypeName, goType.ArrayDepth, g.loadCache)
			for _, obj := range arrObjs {
				g.cached[obj.ID] = obj
			}
			objField.Type.Ref = arrObjs[len(arrObjs)-1].ID
			rootObj.Fields = append(rootObj.Fields, objField)
			continue
		}

		pkgTypePath := fmt.Sprintf("%s.%s", goType.ImportPkgName, goType.TypeName)
		obj := g.loadCache(pkgTypePath)
		objCached := true
		if obj == nil {
			objCached = false
			obj = &mkdoc.Object{
				ID: pkgTypePath,
				Type: &mkdoc.ObjectType{
					Name:       "object",
					Ref:        "",
					IsRepeated: false,
				},
				Fields: nil,
				Loaded: false,
			}
			*queue = append(*queue, pkgTypePath)
		}
		var arrObjs []*mkdoc.Object
		if goType.IsArray {
			arrObjs = mkdoc.CreateArrayObject(obj, goType.ArrayDepth)
		} else {
			arrObjs = mkdoc.CreateArrayObject(obj, 0)
		}

		for _, o := range arrObjs {
			if !objCached {
				g.cached[o.ID] = o
				continue
			}
			if o.ID != obj.ID {
				g.cached[o.ID] = o
			}
		}
		objField.Type.Ref = arrObjs[len(arrObjs)-1].ID

		rootObj.Fields = append(rootObj.Fields, objField)
	}
	rootObj.Loaded = true
	return nil
}

func (g *GoLoader) getStructInfo(query *PkgType) (*GoStructInfo, error) {
	var structInfo *GoStructInfo
	var err error
	if g.enableGoMod {
		pkgAbsPath := strings.Replace(query.Package, g.mod.ModulePkg, g.mod.ModulePath, 1)
		structInfo, err = newStructFinder(g.mod).Find(pkgAbsPath, query.TypeName)
		if err != nil {
			return nil, err
		}
		if structInfo == nil {
			return nil, fmt.Errorf("struct %s not found\n", query)
		}
		return structInfo, nil
	}

	goSrcPaths := mkdoc.GetGOSrcPaths()
	pkgAbsPaths := make([]string, 0)
	for _, p := range goSrcPaths {
		pkgAbsPath := filepath.Join(p, query.Package)
		pkgAbsPaths = append(pkgAbsPaths, pkgAbsPath)
		if _, err := os.Stat(pkgAbsPath); err != nil {
			continue
		}
		structInfo, err = newStructFinder(nil).Find(pkgAbsPath, query.TypeName)
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

func (g *GoLoader) Load(ts mkdoc.TypeScope) (*mkdoc.Object, error) {
	if !g.initialed {
		return nil, errors.New("loader not initialed")
	}
	objs, err := g.LoadAll([]mkdoc.TypeScope{ts})
	if err != nil {
		return nil, err
	}
	if len(objs) == 0 {
		return nil, fmt.Errorf("object not found type scope: %+v", ts)
	}
	return objs[0], nil
}

func (g *GoLoader) SetConfig(cfg *mkdoc.ObjectLoaderConfig) {
	g.init(cfg)
}

func (g *GoLoader) Lang() string {
	return "go"
}
