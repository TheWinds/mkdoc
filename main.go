package main

import "log"

func main() {
	//getAPIDocFuncInfo("corego/service/xyt/api")
	//
	//return
	doc()
	doc1()
	doc2()
}

func doc() {
	inLoc := &TypeLocation{
		PackageName: "corego/service/video/api",
		TypeName:    "GetChannelDataStatisticsReq",
	}
	outLoc := &TypeLocation{
		PackageName: "corego/service/video/api",
		TypeName:    "GetChannelDataStatisticsResp",
	}
	//baseViewLoc := &TypeLocation{
	//	PackageName: "corego/service/xyt/view",
	//	TypeName:    "BaseView",
	//}
	api := NewAPI("[query] channelDataStatistics", "获取频道统计数据", "/channelManage", inLoc, outLoc)
	if err := api.Gen("corego/service/video/api"); err != nil {
		log.Fatal(err)
	}
	api.PrintMarkdown()
	//println(11)
	////api.Print()
	////api.PrintJSON()
	//baseViewObj := new(Object)
	//err := api.getObjectInfo(baseViewLoc, baseViewObj, -1)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//api.setObjectJSONTagAndComment(baseViewObj,nil)
	//err = api.LinkField2Object(baseViewObj,"Data",outLoc.String(),false)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//api.OutArgument = baseViewObj
	//api.PrintJSON(api.OutArgument)
}

func doc1() {
	inLoc := &TypeLocation{
		PackageName: "corego/service/operation/api",
		TypeName:    "TagSearchReq",
	}
	outLoc := &TypeLocation{
		PackageName: "corego/service/operation/api",
		TypeName:    "TagResp",
	}
	//baseViewLoc := &TypeLocation{
	//	PackageName: "corego/service/xyt/view",
	//	TypeName:    "BaseView",
	//}
	api := NewAPI("[query] tags", "获取标签", "/operationManage", inLoc, outLoc)
	if err := api.Gen("corego/service/operation/api"); err != nil {
		log.Fatal(err)
	}
	api.PrintMarkdown()
	//println(11)
	////api.Print()
	////api.PrintJSON()
	//baseViewObj := new(Object)
	//err := api.getObjectInfo(baseViewLoc, baseViewObj, -1)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//api.setObjectJSONTagAndComment(baseViewObj,nil)
	//err = api.LinkField2Object(baseViewObj,"Data",outLoc.String(),false)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//api.OutArgument = baseViewObj
	//api.PrintJSON(api.OutArgument)
}


func doc2() {
	inLoc := &TypeLocation{
		PackageName: "corego/service/operation/api",
		TypeName:    "GetDistributionPartnerGetReq",
	}
	outLoc := &TypeLocation{
		PackageName: "corego/service/operation/api",
		TypeName:    "GetDistributionPartnerGetResp",
	}
	//baseViewLoc := &TypeLocation{
	//	PackageName: "corego/service/xyt/view",
	//	TypeName:    "BaseView",
	//}
	api := NewAPI("[query] getDistributionPartner", "获取分销商", "/operationManage", inLoc, outLoc)
	if err := api.Gen("corego/service/operation/api"); err != nil {
		log.Fatal(err)
	}
	api.PrintMarkdown()
	//println(11)
	////api.Print()
	////api.PrintJSON()
	//baseViewObj := new(Object)
	//err := api.getObjectInfo(baseViewLoc, baseViewObj, -1)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//api.setObjectJSONTagAndComment(baseViewObj,nil)
	//err = api.LinkField2Object(baseViewObj,"Data",outLoc.String(),false)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//api.OutArgument = baseViewObj
	//api.PrintJSON(api.OutArgument)
}
