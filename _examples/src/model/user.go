package model

type User struct {
	// ID
	ID int64 `json:"id"`
	// 用户名
	Name string `json:"name"`
	// 密码
	Password string `json:"pwd"`
	// 年龄
	// 这是年龄字段
	Age     int      `json:"age"`
	Profile *Profile `json:"profile"`
}

type Profile struct {
	Friends []User       `json:"friends"`
	Son     User         `json:"son"`
	Address []Address    `json:"address"`
	TTT     [][]int      `json:"ttt"`
	SSS     [][][]string `json:"sss"`
	Phone   string
}

type Address struct {
	// 代码
	Code int `json:"code"`
	// 详细地址
	Addr string `json:"addr"`
}
