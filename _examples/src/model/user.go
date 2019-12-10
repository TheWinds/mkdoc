package model

type User struct {
	// ID
	ID int64 `json:"id"`
	// 用户名
	Name string `json:"name"`
	// 密码
	Password string `json:"pwd"`
	// 年龄
	Age int `json:"age"`
}
