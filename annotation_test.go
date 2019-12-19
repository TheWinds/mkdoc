package mkdoc

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

type TestGOTyp struct {
	A string
	B int
}

func TestDocAnnotation_ParseToAPI(t *testing.T) {
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
			},
			wantErr: false,
		},
		{
			name: "test fields",
			annotation:
			`@doc 
			 @in fields {
				name string 这是一个Name
				age  int    这是一个Age
			 }`,
			want: &API{
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
				InArgument: &Object{
					Fields: []*ObjectField{
						{
							Name: "name",
							//JSONTag: "name",
							//Comment: "这是一个Name",
							//Type:    "string",
						},
					},
				},
				Query:  map[string]string{},
				Header: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "test go type encoder",
			annotation: DocAnnotation(fmt.Sprintf(
				`
			 @in[json]  type TestGOTyp
			 @out[xml]  type TestGOTyp
			 @loc %s/annotation_test.go:1`, dir)),
			want: &API{
				//InArgumentLoc:  newTypeLocation("docspace.TestGOTyp"),
				InArgEncoder:  "json",
				OutArgEncoder: "xml",
				//OutArgumentLoc: newTypeLocation("docspace.TestGOTyp"),
				Query:       map[string]string{},
				Header:      map[string]string{},
				DocLocation: fmt.Sprintf("%s/annotation_test.go:1", dir),
			},
			wantErr: false,
		},
		{
			name: "test fields encoder",
			annotation:
			`@doc 
			 @in[json] fields {
				name string 这是一个Name
				age  int    这是一个Age
			 }`,
			want: &API{
				InArgument: &Object{
					Fields: []*ObjectField{
						{
							Name: "name",
							//JSONTag: "name",
							//Comment: "这是一个Name",
							//Type:    "string",
						},
						{
							Name: "age",
							//JSONTag: "age",
							//Comment: "这是一个Age",
							//Type:    "int",
						},
					},
				},
				InArgEncoder: "json",
				Query:        map[string]string{},
				Header:       map[string]string{},
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
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.annotation.ParseToAPI()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch tt.name {
			case "test fields", "test fields encoder", "test field type not support":
				if !reflect.DeepEqual(got.InArgument.Fields, tt.want.InArgument.Fields) {
					t.Errorf("ParseToAPI() got = %#v, want %#v", got.InArgument.Fields, tt.want.InArgument.Fields)
				}
				return
			}
			got.Annotation = ""
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseToAPI() got = %#v, want %#v", got, tt.want)
				b, _ := json.MarshalIndent(got, "", "\t")
				fmt.Println(string(b))
				b, _ = json.MarshalIndent(tt.want, "", "\t")
				fmt.Println(string(b))
			}

		})
	}
}
