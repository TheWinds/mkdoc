package view

type BaseView struct {
	// 状态码
	Code int `json:"code"`
	// 提示消息
	Message string `json:"msg"`
	// Data
	Data interface{} `json:"data" doc:"T"`
}
