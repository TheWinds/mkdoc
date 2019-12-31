package objmock

import (
	"fmt"
	"github.com/thewinds/mkdoc"
	"testing"
)

func TestGQLBodyMocker_Mock(t *testing.T) {
	refs := make(map[string]*mkdoc.Object)
	for _, obj := range mkdoc.BuiltinObjects() {
		refs[obj.ID] = obj
	}

	parseTag := func(raw string) *mkdoc.ObjectFieldTag {
		tag, _ := mkdoc.NewObjectFieldTag(raw)
		return tag
	}

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
	o, err := GqlBodyMocker().Mock(user, refs)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(o)

	o, err = GqlBodyMocker().MockPretty(user, refs,"","    ")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(o)
}
