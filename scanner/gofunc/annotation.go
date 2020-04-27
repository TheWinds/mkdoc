package gofunc

import (
	"encoding/json"
	"fmt"
	"github.com/thewinds/mkdoc"
	"github.com/thewinds/mkdoc/schema"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

// DocAnnotation is a set of annotation command
// syntax example:
/*
	@doc 获取用户信息
	根据用户ID获取用户信息,如果用户不存在则返回null
	@tag user,profile
	@path /user/:uid/profile @method POST
	@query uid 用户id
	@in  type view.GetUserReq
	@out type api.User
	@out[json] type api.User
	@out fields {
		id   int    用户id
		name string 用户名
		age  int    年龄
	}
	@disable common_header
	@disable base_type
*/
type DocAnnotation string

var annotationRegexps map[string]*regexp.Regexp

func init() {
	annotationRegexps = map[string]*regexp.Regexp{
		"in_go_type":       regexp.MustCompile(`(@in(\[\S*\])?\s+type\s+)([^\s]+)`),
		"out_go_type":      regexp.MustCompile(`(@out(\[\S*\])?\s+type\s+)([^\s]+)`),
		"in_fields_block":  regexp.MustCompile(`(@in(\[\S*\])?\s+fields\s+(\[\])?{\s+)((.|\s)+)}`),
		"out_fields_block": regexp.MustCompile(`(@out(\[\S*\])?\s+fields\s+(\[\])?{\s+)((.|\s)+)}`),
		"field":            regexp.MustCompile(`(\w+)\s+(\w+)\s*(.+)*`),
	}
}

// ParseToAPI parse doc annotation to API def struct
func (annotation DocAnnotation) ParseToAPI() (*schema.API, []*schema.Object, error) {
	api := new(schema.API)
	api.Language = "go"
	err := parseSimple(annotation, api)
	if err != nil {
		return nil, nil, err
	}
	objects, err := parseInOut(annotation, api)
	if err != nil {
		return nil, nil, err
	}
	return api, objects, nil
}

func parseSimple(annotation DocAnnotation, api *schema.API) error {
	api.Query = make(map[string]string)
	api.Header = make(map[string]string)

	lines := strings.Split(string(annotation), "\n")
	lineFields := make([][]string, 0, len(lines))

	for _, line := range lines {
		lineFields = append(lineFields, strings.Fields(line))
	}
	var lastCmd string
	sbDescription := strings.Builder{}

	for _, fields := range lineFields {
		fieldNum := len(fields)
		if fieldNum == 0 {
			continue
		}
		isCmd := strings.HasPrefix(fields[0], "@")
		if !isCmd {
			if lastCmd == "@doc" {
				if sbDescription.Len() != 0 {
					sbDescription.WriteString("\n")
				}
				sbDescription.WriteString(strings.Join(fields, " "))
			}
			continue
		}

		if isCmd && fieldNum == 1 {
			continue
		}

		cmd := fields[0]
		lastCmd = cmd
		// simple command
		switch cmd {
		case "@doc":
			api.Name = fields[1]
		case "@type":
			api.Type = fields[1]
		case "@method":
			api.Method = fields[1]
		case "@path":
			api.Path = fields[1]
			if fieldNum >= 4 && fields[2] == "@method" {
				api.Method = fields[3]
			}
		case "@tag":
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
		case "@query":
			name := fields[1]
			comment := ""
			if len(fields) >= 3 {
				comment = fields[2]
			}
			api.Query[name] = comment
		case "@header":
			name := fields[1]
			comment := ""
			if len(fields) >= 3 {
				comment = fields[2]
			}
			api.Header[name] = comment
		case "@loc":
			loc := strings.Split(fields[1], ":")
			if len(loc) == 2 {
				api.SourceFileName = loc[0]
				api.SourceLineNum, _ = strconv.Atoi(loc[1])
			}
		case "@disable":
			api.Disables = append(api.Disables, fields[1])
		default:
		}
	}
	api.Desc = sbDescription.String()
	return nil
}

func parseInOut(annotation DocAnnotation, api *schema.API) ([]*schema.Object, error) {
	var objects []*schema.Object
	for command, re := range annotationRegexps {
		matchGroups := re.FindAllStringSubmatch(string(annotation), -1)
		for _, matchGroup := range matchGroups {
			if len(matchGroup) > 0 {
				switch command {
				case "in_go_type":
					api.MimeIn = rmBracket(matchGroup[2])
					api.InType = matchGroup[3]
				case "out_go_type":
					api.MimeOut = rmBracket(matchGroup[2])
					api.OutType = matchGroup[3]
				case "in_fields_block":
					api.MimeIn = rmBracket(matchGroup[2])
					fieldStmts := matchGroup[4]
					obj := &schema.Object{
						ID:     mkdoc.RandObjectID("in"),
						Type:   &schema.ObjectType{Name: "object"},
						Fields: parseToObjectFields(fieldStmts),
					}
					if matchGroup[3] != "" {
						obj.Type.IsRepeated = true
					}
					objects = append(objects, obj)
				case "out_fields_block":
					api.MimeOut = rmBracket(matchGroup[2])
					fieldStmts := matchGroup[4]
					obj := &schema.Object{
						ID:     mkdoc.RandObjectID("out"),
						Type:   &schema.ObjectType{Name: "object"},
						Fields: parseToObjectFields(fieldStmts),
					}
					if matchGroup[3] != "" {
						obj.Type.IsRepeated = true
					}
					objects = append(objects, obj)
				}
			}
		}
	}
	return objects, nil
}

func parseToObjectFields(fieldStmts string) []*schema.ObjectField {
	fields := make([]*schema.ObjectField, 0)
	for _, stmt := range strings.Split(fieldStmts, "\n") {
		matchGroups := annotationRegexps["field"].FindStringSubmatch(stmt)
		if len(matchGroups) > 0 {
			if !isBuiltinType(matchGroups[2]) {
				fmt.Printf("type [%s] is not support,skip\n", matchGroups[2])
				continue
			}
			ext := &schema.Extension{
				Name: "go_tag",
				Data: json.RawMessage(fmt.Sprintf(`"json:\"%s\" xml:\"%s\""`, matchGroups[1], matchGroups[1])),
			}
			fields = append(fields, &schema.ObjectField{
				Name:       matchGroups[1],
				Desc:       strings.TrimSpace(matchGroups[3]),
				Type:       &schema.ObjectType{Name: matchGroups[2]},
				Extensions: []*schema.Extension{ext},
			})
		}
	}
	return fields
}

func rmBracket(s string) string {
	if len(s) <= 2 {
		return ""
	}
	return s[1 : len(s)-1]
}

func (annotation DocAnnotation) AppendMetaData(typ string, fp token.Position) DocAnnotation {
	t := fmt.Sprintf("@type %s\n", typ)
	loc := fmt.Sprintf("@loc %s:%d\n", fp.Filename, fp.Line)
	return annotation + DocAnnotation(t+loc)
}

const annotationDocToken = "@doc"

// GetAnnotationFromDoc get the annotation from comment doc
// if not found any annotation return ""
func GetAnnotationFromComment(s string) DocAnnotation {
	i := strings.LastIndex(s, annotationDocToken)
	if i == -1 {
		return ""
	}
	return DocAnnotation(s[i:])
}
