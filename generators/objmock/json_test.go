package objmock

import (
	"fmt"
	"github.com/thewinds/mkdoc"
	"testing"
)

func TestJSONMocker_Mock(t *testing.T) {
	objUser := &mkdoc.Object{
		ID: "user",
		Fields: []*mkdoc.ObjectField{
			{
				Name:     "ID",
				JSONTag:  "id",
				Comment:  "user id",
				Type:     "",
				BaseType: "int",
			},
			{
				Name:     "Name",
				JSONTag:  "name",
				Comment:  "user name",
				Type:     "",
				BaseType: "string",
			},
			{
				Name:       "Friends",
				JSONTag:    "friends",
				Comment:    "user friends",
				Type:       "user",
				BaseType:   "",
				IsRepeated: true,
				IsRef:      true,
			},
			{
				Name:     "Son",
				JSONTag:  "son",
				Comment:  "user son",
				Type:     "user",
				BaseType: "",
				IsRef:    true,
			},
			{
				Name:     "Age",
				JSONTag:  "age",
				Comment:  "user age",
				Type:     "",
				BaseType: "int",
			},
			{
				Name:    "Computer",
				JSONTag: "computer",
				Comment: "user computer",
				Type:    "computer",
				IsRef:   true,
			},
		},
	}
	objComputer := &mkdoc.Object{
		ID: "computer",
		Fields: []*mkdoc.ObjectField{
			{
				Name:     "Brand",
				JSONTag:  "brand",
				Comment:  "computer brand",
				Type:     "",
				BaseType: "string",
			},
			{
				Name:     "CPU",
				JSONTag:  "cpu",
				Comment:  "computer cpu",
				Type:     "",
				BaseType: "string",
			},
			{
				Name:     "Price",
				JSONTag:  "price",
				Comment:  "computer price",
				Type:     "",
				BaseType: "float",
			},
		},
	}

	refs := map[string]*mkdoc.Object{
		objUser.ID:     objUser,
		objComputer.ID: objComputer,
	}
	mocker := new(JSONMocker)
	o, err := mocker.Mock(objUser, refs)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(o)
}
