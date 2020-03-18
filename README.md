# MKDOC
> 灵活可定制,基于注释注解的API文档生成器

[![asciicast](https://asciinema.org/a/fIDwADlE8X1MtCCSNb8bUJPte.svg)](https://asciinema.org/a/fIDwADlE8X1MtCCSNb8bUJPte)

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
| *docsify* | 生成docsify所需文档                | [🛸](https://github.com/TheWinds/mkdoc/tree/master/generators/docsify) |
| *markdown* | 生成markdown格式的文档                | [🛸](https://github.com/TheWinds/mkdoc/tree/master/generators/markdown) |
| *insomnia* | 生成可供insomnia导入的数据,可用于测试 | [🛸](https://github.com/TheWinds/mkdoc/tree/master/generators/insomnia) |

