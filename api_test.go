package docspace

import "testing"

func TestAPI_getObjectInfoV2(t *testing.T) {
	err:=new(API).getObjectInfoV2(&TypeLocation{
		PackageName: "docspace",
		TypeName:    "API",
		IsRepeated:  false,
	}, nil, 0)

	if err != nil {
		t.Error(err)
		return
	}
}
