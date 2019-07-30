package main

import "log"

func main() {

	inLoc := &TypeLocation{
		PackageName: "corego/service/xyt/view",
		TypeName:    "BaseView",
	}
	outLoc := &TypeLocation{
		PackageName: "corego/service/zhike-teacher/legacyapi",
		TypeName:    "GetTaskListResp",
	}

	api := NewAPI("test", "测试API", "/zhike/test", inLoc, outLoc)
	if err := api.Gen("corego/service/xyt/router"); err != nil {
		log.Fatal(err)
	}
	api.Print()
	api.PrintJSON()

}