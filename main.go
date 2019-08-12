package main

import "fmt"

func main() {
	docSaleLeads()
}

func docSaleLeads() {
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
			"[mutation] classAdvisers",
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
		Name:       "[query] courseSale",
		Comment:    "销售线索搜索",
		RouterPath: "/zhike/courseManage",
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
					Comment:    "查询参数 新增classAdviserId字段 0-公共 非0-个人",
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
		Name:       "[query] userCourseList",
		Comment:    "销售线索/课程列表",
		RouterPath: "/zhike/courseManage",
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
		Name:       "[mutation] materielCommentUpdate",
		Comment:    "更新发货备注",
		RouterPath: "/zhike/courseManage",
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
		Name:       "[mutation] setCustomerSaleTag",
		Comment:    "设置用户的销售标签",
		RouterPath: "/zhike/courseManage",
		inArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendSetCustomerSaleTagReq",
		},
		outArgumentLoc: &TypeLocation{
			PackageName: pkgZkStudent,
			TypeName:    "BackendSetCustomerSaleTagResp",
		},
	}

	apiALLSaleTags := &API{
		Name:       "[query] saleTags",
		Comment:    "获取所有销售标签",
		RouterPath: "/zhike/courseManage",
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
		Name:       "[query] customerSaleTags",
		Comment:    "获取用户的所有销售标签",
		RouterPath: "/zhike/courseManage",
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
		Name:       "[query] saleTagStatistics",
		Comment:    "获取标签统计",
		RouterPath: "/zhike/courseManage",
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
