package src

import (
	"context"
	"github.com/thewinds/mkdoc/example/model"
)

var user *model.User

// @doc 创建用户
// create a user
// @tag user
// @path /api/user @method post
// @in fields {
//   name string 用户名
//   pwd  string 密码
//   age  int    年龄
// }
// @out type model.User
func CreateUser(ctx context.Context) {
	// ...
}

type CreateUserV2Req struct {
	// 用户名
	Name string `json:"name"`
	// 密码
	Password string `json:"pwd"`
	// 年龄
	Age int `json:"age"`
}

// @doc 获取用户
// get user by id
// @tag user
// @path /api/v2/user @method post
// @query uid 用户ID
// @in  type CreateUserV2Req
// @out type model.User
func CreateUserV2() {
	// ...
}

// @doc 搜索用户
// get user by id
// @tag user
// @path /api/user/ @method get
// @query uid  用户ID
// @query age  年龄
// @query name 名称
// @out type []model.User
func GetUsers() {
	// ...
}

// @doc AAA
// get user by id
// @tag user
// @path /api/aaa/ @method get
// @query uid  用户ID
// @query age  年龄
// @query name 名称
// @out type string
func AAA() {
	// ...
}