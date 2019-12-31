package objmock

import (
	"strings"
	"testing"
)

func TestGQLBodyMocker_Mock(t *testing.T) {
	type fields struct {
		refs    map[string]*mkdoc.Object
		err     error
		dep     int
		data    strings.Builder
		refPath []string
	}
	type args struct {
		object *mkdoc.Object
		refs   map[string]*mkdoc.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GQLBodyMocker{
				refs:    tt.fields.refs,
				err:     tt.fields.err,
				dep:     tt.fields.dep,
				data:    tt.fields.data,
				refPath: tt.fields.refPath,
			}
			got, err := g.Mock(tt.args.object, tt.args.refs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Mock() got = %v, want %v", got, tt.want)
			}
		})
	}
}