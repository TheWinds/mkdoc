# docspace
> 从GO源代码直接生成API文档



### 快速开始

```
usage: mkdoc --scanner=SCANNER [<flags>] <pkg> [<out>]

Flags:
      --help             Show context-sensitive help (also try --help-long and --help-man).
  -s, --scanner=SCANNER  which api scanner to use,eg. gql-corego
  -t, --tag=TAG          which tag to filter,eg. v1

Args:
  <pkg>    which package to scan
  [<out>]  which file to output
```



### 文档注解

> 文档注解以注释的形式写在go源码中,不同的扫描器会从不同的位置扫描注解
>
> 例如内建的`funcdoc`扫描器将会扫描所有的方法声明上的文档注解
>
> - 所有注解以 `@apidoc` 开头,目前支持以下注解

| 注解命令 | 说明 |
| ----- | ----- |
|@apidoc `name <name>`|名称|
|@apidoc `desc <desc>`|描述|
|@apidoc `name <name> desc <desc>`|名称+描述|
|@apidoc `type <type>`|类型|
|@apidoc `path <path>`|路径|
|@apidoc `method <method>`|请求方法|
|@apidoc `path <path> method <method>`|路径+请求方法|
|@apidoc `tag <tag>`|标签: 多个以`,`分隔|
|@apidoc `in <params>`|入参类型|
|@apidoc `out <params>`|出参(返回)类型|
其中 `in` 和 `out` 支持两种形式

- 一种是直接根据给定包名和类型名称去引用 GoType ，mkdoc 将会找到Type定义利用其注释信息得出文档所需信息。这种方式支持任意层级的类型嵌套。

```go
// @apidoc in/out gotype package.type

// -- 例子

// main/xx.go
// ...

// @apidoc name getUser
// ...
// @apidoc out gotype model.User
func GetUser(ctx echo.Context){
  // ...
}

// model/user.go
type User struct{
  ID   int    `json:"id"`   // id
  Name string `json:"name"` // 名称
}
```

- 另一种是，是直接写出Type定义,这种方式只支持一层的字段定义。

```go
// @apidoc in/out fields {
//   fieldName filedType comment
//}

// -- 例子
// main/xx.go
// ...

// @apidoc name getUser
// ...
// @apidoc out fields {
//   id   int    id
//   name string 名称
//}
func GetUser(ctx echo.Context){
  // ...
}
```

