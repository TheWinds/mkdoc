package objmock

import (
	"fmt"
	"github.com/thewinds/mkdoc"
	"testing"
)

func TestJSONMocker_Mock(t *testing.T) {
	refs := make(map[string]*mkdoc.Object)
	for _, obj := range mkdoc.BuiltinObjects() {
		refs[obj.ID] = obj
	}

	parseTag := func(raw string) *mkdoc.ObjectFieldTag {
		tag, _ := mkdoc.NewObjectFieldTag(raw)
		return tag
	}

	type want struct {
		objectID string
		wantJSON string
	}

	var wants []want

	refs["test_string"] = &mkdoc.Object{
		ID:   "test_string",
		Type: &mkdoc.ObjectType{Name: "string"},
	}
	wants = append(wants, want{"test_string", `"str"`})

	refs["test_string_arr"] = &mkdoc.Object{
		ID:   "test_string",
		Type: &mkdoc.ObjectType{Name: "string", IsRepeated: true},
	}
	wants = append(wants, want{"test_string_arr", `["str"]`})

	refs["test_int_arr_dep0"] = &mkdoc.Object{
		ID:   "test_int_arr_dep0",
		Type: &mkdoc.ObjectType{Name: "object", IsRepeated: true, Ref: "test_int_arr_dep1"},
	}
	refs["test_int_arr_dep1"] = &mkdoc.Object{
		ID:   "test_int_arr_dep1",
		Type: &mkdoc.ObjectType{Name: "object", IsRepeated: true, Ref: "int"},
	}
	wants = append(wants, want{"test_int_arr_dep0", `[[10]]`})

	profile := &mkdoc.Object{
		ID: "abc.profile",
		Type: &mkdoc.ObjectType{
			Name: "object",
		},
		Fields: []*mkdoc.ObjectField{
			{
				Name: "NickName",
				Desc: "name",
				Type: &mkdoc.ObjectType{Name: "string"},
				Tag:  parseTag(`json:"nickname"`),
			},
			{
				Name: "age",
				Desc: "age",
				Type: &mkdoc.ObjectType{Name: "int"},
				Tag:  parseTag(`json:"age"`),
			},
		},
		Loaded: false,
	}
	refs[profile.ID] = profile

	user := &mkdoc.Object{
		ID: "abc.user",
		Type: &mkdoc.ObjectType{
			Name: "object",
		},
		Fields: []*mkdoc.ObjectField{
			{
				Name: "id",
				Desc: "id",
				Type: &mkdoc.ObjectType{Name: "int64"},
				Tag:  parseTag(`json:"uid"`),
			},
			{
				Name: "onLine",
				Desc: "user name",
				Type: &mkdoc.ObjectType{Name: "bool"},
				Tag:  parseTag(`json:"online"`),
			},
			{
				Name: "profile",
				Desc: "user profile",
				Type: &mkdoc.ObjectType{Name: "object", Ref: "abc.profile"},
				Tag:  parseTag(`json:"profile"`),
			},
			{
				Name: "friends",
				Desc: "user friends",
				Type: &mkdoc.ObjectType{Name: "object", Ref: "abc.user", IsRepeated: true},
				Tag:  parseTag(`json:"friends"`),
			},
		},
		Loaded: false,
	}
	refs[user.ID] = user
	wants = append(wants, want{user.ID, `{"uid":10,"online":true,"profile":{"nickname":"str","age":10},"friends":{"uid":10,"online":true,"profile":{"nickname":"str","age":10},"friends":null}}`})

	for _, w := range wants {
		fmt.Println("Test", w.objectID)
		mocker := new(JSONMocker)
		o, err := mocker.MockNoComment(refs[w.objectID], refs)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println(o)
		if o != w.wantJSON {
			t.Errorf("\n got= %s\n want=%s\n", o, w.wantJSON)
			return
		}
		fmt.Println("Pass")
	}
}
