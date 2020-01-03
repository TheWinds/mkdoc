package mkdoc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

type TestGOTyp struct {
	A string
	B int
}

func TestDocAnnotation_ParseToAPI(t *testing.T) {
	if _project == nil {
		config := &Config{
			Name:         "test annotation",
			Description:  "",
			APIBaseURL:   "",
			Mime:   &MimeType{"json", ""},
			CommonHeader: []*Header{},
			Package:      ".",
			BaseType:     "",
			UseGOModule:  true,
			Scanner:      []string{"funcdoc"},
			Generator:    []string{},
		}
		if err := config.Check(); err != nil {
			t.Error(err)
			return
		}
		_project = &Project{Config: config}

		if config.UseGOModule {
			if err := _project.initGoModule(); err != nil {
				t.Error(err)
				return
			}
		}
		_project.refObjects = make(map[string]*Object)
	}

	dir, _ := os.Getwd()
	tests := []struct {
		name       string
		annotation DocAnnotation
		want       *API
		wantErr    bool
	}{
		{
			name: "basic",
			annotation:
			`@doc abc
			测试API
			@type graphql
			@path /api/v1/abc
			@method query
			@tag v1
			@query uid 用户ID
			@query pwd 密码
			@header token  jwtToken
			@header userId userId`,
			want: &API{
				Name:   "abc",
				Desc:   "测试API",
				Path:   "/api/v1/abc",
				Method: "query",
				Type:   "graphql",
				Tags:   []string{"v1"},
				Query:  map[string]string{"uid": "用户ID", "pwd": "密码"},
				Header: map[string]string{"token": "jwtToken", "userId": "userId"},
			},
			wantErr: false,
		},
		{
			name: "command combine",
			annotation:
			`@doc abc
			测试API
			@type graphql
			@path /api/v1/abc @method query
			@tag v1`,
			want: &API{
				Name:   "abc",
				Desc:   "测试API",
				Path:   "/api/v1/abc",
				Method: "query",
				Type:   "graphql",
				Tags:   []string{"v1"},
				Query:  map[string]string{},
				Header: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "multi tag",
			annotation:
			`@doc abc
			测试API
			@type graphql
			@path /api/v1/abc
			@method query
			@tag v1,test`,
			want: &API{
				Name:   "abc",
				Desc:   "测试API",
				Path:   "/api/v1/abc",
				Method: "query",
				Type:   "graphql",
				Tags:   []string{"v1", "test"},
				Query:  map[string]string{},
				Header: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "test go type",
			annotation: DocAnnotation(fmt.Sprintf(
				`
			 @in  type TestGOTyp
			 @out type TestGOTyp
			 @loc %s/annotation_test.go:1`, dir)),
			want: &API{
				Query:       map[string]string{},
				Header:      map[string]string{},
				DocLocation: fmt.Sprintf("%s/annotation_test.go:1", dir),
				InArgument: &Object{
					ID:     "github.com/thewinds/mkdoc.TestGOTyp",
					Type:   nil,
					Fields: nil,
					Loaded: false,
				},
				OutArgument: &Object{
					ID:     "github.com/thewinds/mkdoc.TestGOTyp",
					Type:   nil,
					Fields: nil,
					Loaded: false,
				},
			},
			wantErr: false,
		},
		{
			name: "test fields",
			annotation: DocAnnotation(fmt.Sprintf(
				`@doc 
			 @in fields {
				name string 这是一个Name
				age  int    这是一个Age
			 }
			 @loc %s/annotation_test.go:1`, dir)),
			want: &API{
				DocLocation: fmt.Sprintf("%s/annotation_test.go:1", dir),
				InArgument: &Object{
					Fields: []*ObjectField{
						{
							Name: "name",
							Tag:  mustObjectFieldTag(`json:"name" xml:"name"`),
							Desc: "这是一个Name",
							Type: &ObjectType{Name: "string"},
						},
						{
							Name: "age",
							Tag:  mustObjectFieldTag(`json:"age" xml:"age"`),
							Desc: "这是一个Age",
							Type: &ObjectType{Name: "int"},
						},
					},
					Loaded: true,
				},
				Query:  map[string]string{},
				Header: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "test field type not support",
			annotation:
			`@doc
			 @in[json] fields {
				name string 这是一个Name
				age  int11    这是一个Age
			 }`,
			want: &API{
				Query:  map[string]string{},
				Header: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "test go type encoder",
			annotation: DocAnnotation(fmt.Sprintf(
				`
			 @in[json]  type TestGOTyp
			 @out[xml]  type TestGOTyp
			 @loc %s/annotation_test.go:1`, dir)),
			want: &API{
				Mime:        &MimeType{"json", "xml"},
				Query:       map[string]string{},
				Header:      map[string]string{},
				DocLocation: fmt.Sprintf("%s/annotation_test.go:1", dir),
				InArgument: &Object{
					ID:     "mkdoc.TestGOTyp",
					Type:   nil,
					Fields: nil,
					Loaded: false,
				},
				OutArgument: &Object{
					ID:     "mkdoc.TestGOTyp",
					Type:   nil,
					Fields: nil,
					Loaded: false,
				},
			},
			wantErr: false,
		},
		{
			name: "test fields encoder",
			annotation:
			DocAnnotation(fmt.Sprintf(
				`@doc 
			 @in[json] fields {
				name string 这是一个Name
				age  int    这是一个Age
			 }
			 @loc %s/annotation_test.go:1`, dir)),
			want: &API{
				DocLocation: fmt.Sprintf("%s/annotation_test.go:1", dir),
				InArgument: &Object{
					Loaded: true,
					Fields: []*ObjectField{
						{
							Name: "name",
							Tag:  mustObjectFieldTag(`json:"name" xml:"name"`),
							Desc: "这是一个Name",
							Type: &ObjectType{Name: "string"},
						},
						{
							Name: "age",
							Tag:  mustObjectFieldTag(`json:"age" xml:"age"`),
							Desc: "这是一个Age",
							Type: &ObjectType{Name: "int"},
						},
					},
				},
				Mime:   &MimeType{"json", ""},
				Query:  map[string]string{},
				Header: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "test disable",
			annotation:
			`@doc 
			 @disable common_header
			 @disable base_type
			 `,
			want: &API{
				Query:    map[string]string{},
				Header:   map[string]string{},
				Disables: []string{"common_header", "base_type"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		ok := t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("TEST %s...\n", tt.name)
			got, err := tt.annotation.ParseToAPI()
			if err != nil {
				if tt.wantErr {
					fmt.Println("PASS", err)
					return
				}
				fmt.Printf("ParseToAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got.Annotation = ""
			if got.InArgument != nil {
				if got.InArgument.ID != "" {
					tt.want.InArgument.ID = got.InArgument.ID
				}
			}
			if got.OutArgument != nil {
				if got.OutArgument.ID != "" {
					tt.want.OutArgument.ID = got.OutArgument.ID
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				b1, _ := json.MarshalIndent(got, "", "    ")
				b2, _ := json.MarshalIndent(tt.want, "", "    ")
				s1, s2 := string(b1), string(b2)
				fmt.Println("GOT:\n", s1)
				fmt.Println("WANT:\n", s2)
				fmt.Printf("DIFF:\n%s", diff(s1, s2))
				t.Fail()
				return
			}
			fmt.Println("PASS")
		})
		if !ok {
			break
		}
	}
}

func diff(s1, s2 string) string {
	tmpdir := os.TempDir()
	f1 := filepath.Join(tmpdir, "diff_f1")
	f2 := filepath.Join(tmpdir, "diff_f2")
	ioutil.WriteFile(f1, []byte(s1), 0644)
	ioutil.WriteFile(f2, []byte(s2), 0644)
	cmd := exec.Command("git", "diff", f1, f2)
	b, err := cmd.Output()
	if err != nil {
		if err.(*exec.ExitError).ExitCode() != 1 {
			log.Fatal(err)
		}
	}
	os.Remove(f1)
	os.Remove(f2)
	return string(b)
}
