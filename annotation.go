package docspace

import (
	"fmt"
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
	}
}

// ParseToAPI parse doc annotation to API def struct
func (annotation DocAnnotation) ParseToAPI() (*API, error) {
	api := new(API)
	api.Query = make(map[string]string)

	matchPkgGroups := annotationRegexps["pkg_map"].FindAllStringSubmatch(string(annotation), -1)
	pkgMap := make(map[string]string, len(matchPkgGroups))
	for _, matchGroup := range matchPkgGroups {
		if len(matchGroup) > 0 {
			pkgName := matchGroup[2]
			pkgPath := matchGroup[3]
			pkgMap[pkgName] = pkgPath
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
					loc := newTypeLocation(matchGroup[2])
					if pkgMap[loc.PackageName] != "" {
						loc.PackageName = pkgMap[loc.PackageName]
					}
					api.InArgumentLoc = loc
				case "out_gotype":
					loc := newTypeLocation(matchGroup[2])
					if pkgMap[loc.PackageName] != "" {
						loc.PackageName = pkgMap[loc.PackageName]
					}
					api.OutArgumentLoc = loc
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
				}
			}
		}
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
