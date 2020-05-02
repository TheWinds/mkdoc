# MKDOC
> 灵活可定制,多语言支持的API文档生成器

## 快速开始

- 安装

```shell
GO111MODULE=on go get github.com/thewinds/mkdoc/cmd/mkdoc
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

## 文档
[👉 Wiki](https://github.com/TheWinds/mkdoc/wiki)

## 插件
插件包括3种类型*Scanner*、*Generator*、*ObjectLoader*,您可以自己实现这3种插件来适应自己项目中的生成需求,
下面有一些已经实现的插件。
### Scanner

*Scanner*(扫描器)的作用是从源码中扫描注解

内置了以下扫描器:

| 名称    | 说明                      | 链接                                                         |
| ------- | ------------------------- | ------------------------------------------------------------ |
| *gofuc* | 从 golang func comments中扫描文档信息 | [🛸](https://github.com/TheWinds/mkdoc/tree/master/scanner/gofunc) |
| *docdef* | 从 doc schema文件中扫描文档信息 | [🛸](https://github.com/TheWinds/mkdoc/tree/master/scanner/docdef) |



### Generator

*Generator*(文档生成器)的作用是根据api信息生成 文档 || 测试

内置了以下生成器:

| 名称     | 说明                                  | 链接                                  |
| -------- | ------------------------------------- | --------------------------------------- |
| *docsify* | 生成docsify所需文档                | [🛸](https://github.com/TheWinds/mkdoc/tree/master/generator/docsify) |
| *markdown* | 生成markdown格式的文档                | [🛸](https://github.com/TheWinds/mkdoc/tree/master/generator/markdown) |
| *insomnia* | 生成可供insomnia导入的数据,可用于测试 | [🛸](https://github.com/TheWinds/mkdoc/tree/master/generator/insomnia) |

### ObjectLoader
*ObjectLoader*(Object加载器)的作用是根据类型定位信息加载Object

内置了以下Loader:

| 名称     | 说明                                  | 链接                                  |
| -------- | ------------------------------------- | --------------------------------------- |
| *goloader* | golang sturct 类型加载               | [🛸](https://github.com/TheWinds/mkdoc/tree/master/objloader/goloader) |