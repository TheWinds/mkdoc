package main

import (
	"fmt"
)

func main() {
	scanGraphQLAPIDocInfo("corego/service/boss/schemas")
	return
	docSaleLeads()
}

func docSaleLeads() {
	fmt.Println("# SaleLeadsAPI")
	fmt.Println("[TOC]")

	fmt.Println("## 群码相关")
	pkgOperation := "corego/service/operation/api"
	apis := []*API{
		NewAPI(

			"[query] classAdvisers",
			"获取所有班主任",
			"/operationManage",
			&TypeLocation{
				PackageName: pkgOperation,
				TypeName:    "GetALLClassAdvisersReq",
			},
			&TypeLocation{
				PackageName: pkgOperation,
				TypeName:    "OPClassAdviser",
				IsRepeated:  true,
			},
		),
		NewAPI(
			"[mutation] setGroupClassAdviser",
			"分配群班主任",
			"/operationManage",
			&TypeLocation{
				PackageName: pkgOperation,
				TypeName:    "SetGroupClassAdviserReq",
			},
			&TypeLocation{
				PackageName: pkgOperation,
				TypeName:    "SetGroupClassAdviserResp",
			},
		),
		NewAPI(
			"[mutation] groupQRCodeList",
			"获取群码列表",
			"/operationManage",
			&TypeLocation{
				PackageName: pkgOperation,
				TypeName:    "GetGroupQRCodeListReq",
			},
			&TypeLocation{
				PackageName: pkgOperation,
				TypeName:    "GroupQRCode",
				IsRepeated:  true,
			},
		),
	}
	for _, v := range apis {
		v.Gen(pkgOperation)
		v.PrintMarkdown()
		fmt.Println()
	}

	fmt.Println("## 销售线索相关")

	pkgZkStudent := "corego/service/zhike-student/api"

	apiCourseSale := &API{
		Name: "[query] courseSale",
		Desc: "销售线索搜索",
		Path: "/zhike/courseManage",
		InArgument: &Object{
			ID: "courseSale",
			Fields: []*ObjectField{
				{
					Name:       "current",
					JSONTag:    "current",
					Comment:    "当前页",
					Type:       "int",
					IsRepeated: false,
					IsRef:      false,
				},
				{
					Name:       "pageSize",
					JSONTag:    "pageSize",
					Comment:    "页面大小",
					Type:       "int",
					IsRepeated: false,
					IsRef:      false,
				},
				{
					Name:       "query",
					JSONTag:    "query",
					Comment:    "查询参数 新增classAdviserId字段 \"\"-公共 非\"\"-个人",
					Type:       "string",
					IsRepeated: false,
					IsRef:      false,
				},
			},
		},
		OutArgument: &Object{
			ID:     "courseSaleResult",
			Fields: []*ObjectField{},
		},
	}

	apiUserCourseList := &API{
		Name: "[query] userCourseList",
		Desc: "销售线索/课程列表",
		Path: "/zhike/courseManage",
		InArgument: &Object{
			ID: "courseSale",
			Fields: []*ObjectField{
				{
					Name:       "current",
					JSONTag:    "current",
					Comment:    "当前页",
					Type:       "int",
					IsRepeated: false,
					IsRef:      false,
				},
				{
					Name:       "pageSize",
					JSONTag:    "pageSize",
					Comment:    "页面大小",
					Type:       "int",
					IsRepeated: false,
					IsRef:      false,
				},
				{
					Name:       "customerId",
					JSONTag:    "customerId",
					Comment:    "用户ID",
					Type:       "int",
					IsRepeated: false,
					IsRef:      false,
				},
			},
		},
		outArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendUserCourseListEntity",
			IsRepeated:  true,
		},
	}
	apiMaterielCommentUpdate := &API{
		Name: "[mutation] materielCommentUpdate",
		Desc: "更新发货备注",
		Path: "/zhike/courseManage",
		inArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendMaterialCommentLogUpdateRequest",
		},
		outArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendMaterialCommentLogUpdateReply",
		},
	}

	apiSetCustomerSaleTag := &API{
		Name: "[mutation] setCustomerSaleTag",
		Desc: "设置用户的销售标签",
		Path: "/zhike/courseManage",
		inArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendSetCustomerSaleTagReq",
		},
		outArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendSetCustomerSaleTagResp",
		},
	}

	apiDelCustomerSaleTag := &API{
		Name: "[mutation] deleteCustomerSaleTag",
		Desc: "删除用户的销售标签",
		Path: "/zhike/courseManage",
		inArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendDeleteCustomerSaleTagReq",
		},
		outArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendDeleteCustomerSaleTagResp",
		},
	}

	apiALLSaleTags := &API{
		Name: "[query] saleTags",
		Desc: "获取所有销售标签",
		Path: "/zhike/courseManage",
		inArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendGetALLSaleTagsReq",
		},
		outArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "SaleTag",
			IsRepeated:  true,
		},
	}

	apiCustomerSaleTags := &API{
		Name: "[query] customerSaleTags",
		Desc: "获取用户的所有销售标签",
		Path: "/zhike/courseManage",
		inArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendGetCustomerSaleTagsReq",
		},
		outArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "SaleTag",
			IsRepeated:  true,
		},
	}

	apiSaleTagStatistics := &API{
		Name: "[query] saleTagStatistics",
		Desc: "获取标签统计",
		Path: "/zhike/courseManage",
		inArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendGetTagStatisticsReq",
		},
		InArgument: &Object{
			ID: "saleTagStatisticsResp",
			Fields: []*ObjectField{
				{
					Name:       "rows",
					JSONTag:    "rows",
					Comment:    "表格的行集合",
					Type:       "[][]string",
					IsRepeated: false,
					IsRef:      false,
				},
			},
		},
	}

	apis = []*API{
		apiCourseSale,
		apiUserCourseList,
		apiMaterielCommentUpdate,
		apiSetCustomerSaleTag,
		apiDelCustomerSaleTag,
		apiALLSaleTags,
		apiCustomerSaleTags,
		apiSaleTagStatistics,
	}
	for _, v := range apis {
		v.Gen(pkgZkStudent)
		v.PrintMarkdown()
		fmt.Println()
	}
}
