# docspace
> 从GO源代码直接快速生成API文档

## 快速开始

### 生成命令

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
所有注解以 `@apidoc` 开头,支持以下注解
| 命令| 作用 |
| ----- | ----- |
|@apidoc `name <name>`|API名称|
|@apidoc `desc <desc>`|API描述|
|@apidoc `name <name>`|API名称|
|@apidoc `name <name> desc <desc>`|API名称+描述|
...