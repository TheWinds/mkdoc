package docspace

import (
	"fmt"
	"go/token"
	"math/rand"
	"regexp"
	"strings"
)

// DocAnnotation is a set of annotation command
type DocAnnotation string

var annotationRegexps map[string]*regexp.Regexp

func init() {
	annotationRegexps = map[string]*regexp.Regexp{
		"name":             regexp.MustCompile(`(@apidoc\s+name\s+)([^\s]+)`),
		"desc":             regexp.MustCompile(`(@apidoc\s+desc\s+)([^\s]+)`),
		"name_desc":        regexp.MustCompile(`(@apidoc\s+name\s+)([^\s]+)(\s+desc\s+)([^\s]+)`),
		"type":             regexp.MustCompile(`(@apidoc\s+type\s+)([^\s]+)`),
		"path":             regexp.MustCompile(`(@apidoc\s+path\s+)([^\s]+)`),
		"method":           regexp.MustCompile(`(@apidoc\s+method\s+)([^\s]+)`),
		"path_method":      regexp.MustCompile(`(@apidoc\s+path\s+)([^\s]+)(\s+method\s+)([^\s]+)`),
		"tag":              regexp.MustCompile(`(@apidoc\s+tag\s+)([^\s]+)`),
		"in_gotype":        regexp.MustCompile(`(@apidoc\s+in\s+gotype\s+)([^\s]+)`),
		"out_gotype":       regexp.MustCompile(`(@apidoc\s+out\s+gotype\s+)([^\s]+)`),
		"in_fileds_block":  regexp.MustCompile(`(@apidoc\s+in\s+fields\s+(\[\])?{\s+)((.|\s)+)}`),
		"out_fileds_block": regexp.MustCompile(`(@apidoc\s+out\s+fields\s+(\[\])?{\s+)((.|\s)+)}`),
		"field":            regexp.MustCompile(`(\w+)\s+(\w+)\s*(.+)*`),
		"pkg_map":          regexp.MustCompile(`(@apidoc\s+pkg_map\s+)([^\s]+)\s+([^\s]+)`),
		"query":            regexp.MustCompile(`(@apidoc\s+query\s+)([^\s]+)\s+([^\s]+)`),
		"header":           regexp.MustCompile(`(@apidoc\s+header\s+)([^\s]+)\s+([^\s]+)`),
		"loc":              regexp.MustCompile(`(@apidoc\s+loc\s+)([^\s]+)`),
	}
}

// ParseToAPI parse doc annotation to API def struct
func (annotation DocAnnotation) ParseToAPI() (*API, error) {
	api := new(API)
	api.Query = make(map[string]string)
	api.Header = make(map[string]string)

	lines := strings.Split(string(annotation), "\n")
	lineFields := make([][]string, 0, len(lines))

	for _, line := range lines {
		lineFields = append(lineFields, strings.Fields(line))
	}
	/*
	@doc
	name xxxxx
	desc sdsdsdssds
	path /a/b/c
	method xxxx

	 */
	// @doc
	// 获取用户
	// 撒旦所阿萨德
	// @type
	// @name user
	// @path /a/b method post
	// @in[json] fields {
	// a int 哈哈
	// }
	// out[xml] go_type api.User
	for _, fields := range lineFields {
		fieldNum := len(fields)
		if fieldNum == 0 {
			continue
		}
		if fieldNum == 1 {

		}
		cmd := fields[0]
		switch cmd {
		case "name":
			api.Name = fields[1]
		case "desc":
			api.Desc = fields[1]
		case "name_desc":
			api.Name = matchGroup[2]
			api.Desc = matchGroup[4]
		case "type":
			api.Type = fields[1]
		case "path":
			api.Path = fields[1]
		case "method":
			api.Method = fields[1]
		case "path_method":
			api.Path = matchGroup[2]
			api.Method = matchGroup[4]
		case "tag":
			tagsStr := fields[1]
			api.Tags = make([]string, 0)
			if strings.Contains(tagsStr, ",") {
				for _, tag := range strings.Split(tagsStr, ",") {
					if tag != "" {
						api.Tags = append(api.Tags, strings.TrimSpace(tag))
					}
				}
			} else {
				api.Tags = append(api.Tags, strings.TrimSpace(tagsStr))
			}
		case "in_gotype":
			api.InArgumentLoc = newTypeLocation(fields[1])
		case "out_gotype":
			api.OutArgumentLoc = newTypeLocation(fields[1])
		case "in_fileds_block":
			// TODO: isRepeated := matchGroup[2] != ""
			fieldStmts := matchGroup[3]
			api.InArgument = &Object{
				ID:     fmt.Sprintf("#obj_%d", rand.Int63()),
				Fields: parseToObjectFields(fieldStmts),
			}
		case "out_fileds_block":
			fieldStmts := matchGroup[3]
			api.OutArgument = &Object{
				ID:     fmt.Sprintf("#obj_%d", rand.Int63()),
				Fields: parseToObjectFields(fieldStmts),
			}
		case "query":
			queryName := fields[1]
			queryComment := fields[2]
			api.Query[queryName] = queryComment
		case "header":
			name := fields[1]
			comment := fields[2]
			api.Header[name] = comment
		case "loc":
			api.DocLocation = fields[1]
		}
	}

}

for command, re := range annotationRegexps {
matchGroups := re.FindAllStringSubmatch(string(annotation), -1)
for _, matchGroup := range matchGroups {
if len(matchGroup) > 0 {
switch command {
case "name":
api.Name = matchGroup[2]
case "desc":
api.Desc = matchGroup[2]
case "name_desc":
api.Name = matchGroup[2]
api.Desc = matchGroup[4]
case "type":
api.Type = matchGroup[2]
case "path":
api.Path = matchGroup[2]
case "method":
api.Method = matchGroup[2]
case "path_method":
api.Path = matchGroup[2]
api.Method = matchGroup[4]
case "tag":
tagsStr := matchGroup[2]
api.Tags = make([]string, 0)
if strings.Contains(tagsStr, ",") {
for _, tag := range strings.Split(tagsStr, ",") {
if tag != "" {
api.Tags = append(api.Tags, strings.TrimSpace(tag))
}
}
} else {
api.Tags = append(api.Tags, strings.TrimSpace(tagsStr))
}
case "in_gotype":
api.InArgumentLoc = newTypeLocation(matchGroup[2])
case "out_gotype":
api.OutArgumentLoc = newTypeLocation(matchGroup[2])
case "in_fileds_block":
// TODO: isRepeated := matchGroup[2] != ""
fieldStmts := matchGroup[3]
api.InArgument = &Object{
ID:     fmt.Sprintf("#obj_%d", rand.Int63()),
Fields: parseToObjectFields(fieldStmts),
}
case "out_fileds_block":
fieldStmts := matchGroup[3]
api.OutArgument = &Object{
ID:     fmt.Sprintf("#obj_%d", rand.Int63()),
Fields: parseToObjectFields(fieldStmts),
}
case "query":
queryName := matchGroup[2]
queryComment := matchGroup[3]
api.Query[queryName] = queryComment
case "header":
name := matchGroup[2]
comment := matchGroup[3]
api.Header[name] = comment
case "loc":
api.DocLocation = matchGroup[2]
}
}
}
}

// replace package name
fl := strings.Split(api.DocLocation, ":")
if len(fl) != 2 {
fmt.Printf("WARNING: DocLocation is incomplete, got %s\n", api.DocLocation)
return api, nil
}
fileName := fl[0]
imports, err := getFileImportsAtFile(fileName)
if err != nil {
return nil, err
}
if api.InArgumentLoc != nil && api.InArgumentLoc.PackageName != "" {
api.InArgumentLoc.PackageName = imports[api.InArgumentLoc.PackageName]
}
if api.OutArgumentLoc != nil && api.OutArgumentLoc.PackageName != "" {
api.OutArgumentLoc.PackageName = imports[api.OutArgumentLoc.PackageName]
}
return api, nil
}

func parseToObjectFields(fieldStmts string) []*ObjectField {
	fields := make([]*ObjectField, 0)
	for _, stmt := range strings.Split(fieldStmts, "\n") {
		matchGroups := annotationRegexps["field"].FindStringSubmatch(stmt)
		if len(matchGroups) > 0 {
			if !isBuiltinType(matchGroups[2]) {
				fmt.Printf("type [%s] is not support,skip\n", matchGroups[2])
				continue
			}

			fields = append(fields, &ObjectField{
				Name:       matchGroups[1],
				JSONTag:    matchGroups[1],
				Comment:    strings.TrimSpace(matchGroups[3]),
				Type:       matchGroups[2],
				IsRepeated: false,
				IsRef:      false,
			})
		}
	}
	return fields
}

func (annotation DocAnnotation) AppendMetaData(typ string, fp token.Position) DocAnnotation {
	t := fmt.Sprintf("%s type %s\n", annotationDocToken, typ)
	loc := fmt.Sprintf("%s loc %s:%d\n", annotationDocToken, fp.Filename, fp.Column)
	return annotation + DocAnnotation(t+loc)
}

const annotationDocToken = "@apidoc"

// GetAnnotationFromDoc get the annotation from comment doc
// if not found any annotation return ""
func GetAnnotationFromComment(s string) DocAnnotation {
	i := strings.LastIndex(s, annotationDocToken)
	if i == -1 {
		return ""
	}
	return DocAnnotation(s[i+len(annotationDocToken):])
}
