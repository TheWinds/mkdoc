# MKDOC
> 灵活可定制,基于注释注解的API文档生成器

[![asciicast](https://asciinema.org/a/fIDwADlE8X1MtCCSNb8bUJPte.svg)](https://asciinema.org/a/fIDwADlE8X1MtCCSNb8bUJPte)

## 快速开始

- 安装

```shell
GO111MODULE=on go get github.com/TheWinds/mkdoc/cmd/mkdoc
```

- 使用

```bash
cd /path/to/your/projet
# 初始化
mkdoc init
# 修改配置
vim conf.yaml
# 代码注解
# ...
# 生成文档
mkdoc make
```

## 例子
参考[examples](https://github.com/TheWinds/mkdoc/tree/master/_examples)目录下的例子

## 插件
插件包括两种类型*Scanner*和*Generator*,您可以自己实现这两种插件来适应自己项目中的文档需求,
下面有一些已经实现的插件。
### Scanner

*Scanner*(注解扫描器)的作用是从go源码中扫描注解,现在支持以下扫描器:

| 名称    | 说明                      | 链接                                                         |
| ------- | ------------------------- | ------------------------------------------------------------ |
| *funcdoc* | 从func document中获取注解 | [🛸](https://github.com/TheWinds/mkdoc/tree/master/scanners/funcdoc) |



### Generator

*Generator*(文档生成器)的作用是根据api信息生成文档,现在支持以下生成器:

| 名称     | 说明                                  | 链接                                  |
| -------- | ------------------------------------- | --------------------------------------- |
| *markdown* | 生成markdown格式的文档                | [🛸](https://github.com/TheWinds/mkdoc/tree/master/generators/markdown) |
| *insomnia* | 生成可供insomnia导入的数据,可用于测试 | [🛸](https://github.com/TheWinds/mkdoc/tree/master/generators/insomnia) |



## 文档

### 注解
> 文档注解以注释的形式写在go源码中,不同的扫描器会从不同的位置扫描注解
>
> 例如内建的`funcdoc`扫描器将会扫描所有的方法声明上的文档注解
>
> - 所有注解以 `@doc` 开头,目前支持以下注解
> - @doc 到下一个命令之间的内容为文档描述

| 注解命令 | 说明 |
| ----- | ----- |
|`@doc` <name\> |名称 *文档注解起始标志*|
|`@type` <type\>|类型|
|`@path` <path\>|路径|
|`@method` <method\>|请求方法|
|`@path`  <path\> @method <method\>|路径+请求方法|
|`@tag` <tag\>|标签: 多个以,分隔|
|`@header` <header 名称\> <header 说明\>|header信息,多个可重复写|
|`@query`  <query 名称\> <query 说明\>|query信息,多个可重复写|
|`@in` <params\>|入参类型|
|`@out` <params\>|出参(返回)类型|
|`@in[mime_type]`  <params\>|指定mime_type,form,json,xml...|
|`@out[mime_type]` <params\>|指定mime_type,form,json,xml...|

> `in` 和 `out` 后的 `[mime_type]` 例如如果入参类型是通过json格式传递过来的
则可以写`@in[json] xxxx`,xml 则写 `@in[xml] xxx` 

其中 `in` 和 `out` 支持两种形式

- 一种是直接根据给定包名和类型名称去引用 GoType ，mkdoc 将会找到Type定义利用其注释信息得出文档所需信息。这种方式支持任意层级的类型嵌套。

```go
// @doc name
// @in/@out type package.type

// -- 例子

// main/xx.go
// ...

// @doc getUser
// 获取用户
// ...
// @out type model.User
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
// @doc name
// @in/@out fields {
//   fieldName filedType comment
//}

// -- 例子
// main/xx.go
// ...

// @doc getUser
// ...
// @out fields {
//   id   int    id
//   name string 名称
//}
func GetUser(ctx echo.Context){
  // ...
}
```
