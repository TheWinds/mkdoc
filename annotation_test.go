package docspace

import (
	"reflect"
	"testing"
)

type TestGOTyp struct {
	A string
	B int
}

func TestDocAnnotation_ParseToAPI(t *testing.T) {
	tests := []struct {
		name       string
		annotation DocAnnotation
		want       *API
		wantErr    bool
	}{
		{
			name: "basic",
			annotation:
			`@apidoc name abc
			@apidoc  desc 测试API
			@apidoc  type graphql
			@apidoc  path /api/v1/abc
			@apidoc  method query
			@apidoc  tag v1`,
			want: &API{
				Name:   "abc",
				Desc:   "测试API",
				Path:   "/api/v1/abc",
				Method: "query",
				Type:   "graphql",
				Tags:   []string{"v1"},
			},
			wantErr: false,
		},
		{
			name: "command combine",
			annotation:
			`@apidoc name abc desc 测试API
			@apidoc  type graphql
			@apidoc  path /api/v1/abc method query
			@apidoc  tag v1`,
			want: &API{
				Name:   "abc",
				Desc:   "测试API",
				Path:   "/api/v1/abc",
				Method: "query",
				Type:   "graphql",
				Tags:   []string{"v1"},
			},
			wantErr: false,
		},
		{
			name: "multi tag",
			annotation:
			`@apidoc name abc
			@apidoc  desc 测试API
			@apidoc  type graphql
			@apidoc  path /api/v1/abc
			@apidoc  method query
			@apidoc  tag v1,test`,
			want: &API{
				Name:   "abc",
				Desc:   "测试API",
				Path:   "/api/v1/abc",
				Method: "query",
				Type:   "graphql",
				Tags:   []string{"v1", "test"},
			},
			wantErr: false,
		},
		{
			name: "test gotype",
			annotation:
			`@apidoc in gotype docspace.TestGOTyp
			 @apidoc out gotype docspace.TestGOTyp`,
			want: &API{
				inArgumentLoc:  newTypeLocation("docspace.TestGOTyp"),
				outArgumentLoc: newTypeLocation("docspace.TestGOTyp"),
			},
			wantErr: false,
		},
		{
			name: "test fields",
			annotation:
			`@apidoc in fields {
				name string 这是一个Name
				age  int    这是一个Age
			 }`,
			want: &API{
				InArgument: &Object{
					Fields: []*ObjectField{
						{
							Name:    "name",
							JSONTag: "name",
							Comment: "这是一个Name",
							Type:    "string",
						},
						{
							Name:    "age",
							JSONTag: "age",
							Comment: "这是一个Age",
							Type:    "int",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test field type not support",
			annotation:
			`@apidoc in fields {
				name string 这是一个Name
				age  int11    这是一个Age
			 }`,
			want: &API{
				InArgument: &Object{
					Fields: []*ObjectField{
						{
							Name:    "name",
							JSONTag: "name",
							Comment: "这是一个Name",
							Type:    "string",
						},
					},
				},
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
			if tt.name == "test fields" || tt.name == "test field type not support" {
				if !reflect.DeepEqual(got.InArgument.Fields, tt.want.InArgument.Fields) {
					t.Errorf("ParseToAPI() got = %#v, want %#v", got.InArgument.Fields, tt.want.InArgument.Fields)
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseToAPI() got = %#v, want %#v", got, tt.want)
				}
			}

		})
	}
}