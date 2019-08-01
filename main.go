package main

import (
	"log"
)

func main() {
	getAPIDocFuncInfo("corego/service/xyt/api")

	return
	inLoc := &TypeLocation{
		PackageName: "corego/service/xyt/view",
		TypeName:    "BaseView",
	}
	outLoc := &TypeLocation{
		PackageName: "corego/service/xyt/view",
		TypeName:    "TakeCourseView",
	}
	baseViewLoc := &TypeLocation{
		PackageName: "corego/service/xyt/view",
		TypeName:    "BaseView",
	}
	api := NewAPI("test", "测试API", "/zhike/test", inLoc, outLoc)
	if err := api.Gen("corego/service/xyt/router"); err != nil {
		log.Fatal(err)
	}
	//api.Print()
	//api.PrintJSON()
	baseViewObj := new(Object)
	err := api.getObjectInfo(baseViewLoc, baseViewObj, -1)
	if err != nil {
		log.Fatal(err)
	}
	api.setObjectJSONTagAndComment(baseViewObj,nil)
	err = api.LinkField2Object(baseViewObj,"Data",outLoc.String(),false)
	if err != nil {
		log.Fatal(err)
	}
	api.OutArgument = baseViewObj
	api.PrintJSON(api.OutArgument)
}
