# docserver
> clone repository & generate & show doc

## Quick Start
```shell script
# 1.pull
docker pull thewinds/mkdoc-server
# 2.set Environments and Config file
# 3.run
# you will see:
# 2020/05/16 18:48:47 server docs:
# 2020/05/16 18:48:47     index   =>      127.0.0.1:8080
# 2020/05/16 18:48:47     project_0 =>      127.0.0.1:8080/project_0
# 2020/05/16 18:48:47     project_1  =>      127.0.0.1:8080/project_1
# 2020/05/16 18:48:47 notify url: 127.0.0.1:8080/notify
```

> web listen port is `:8080`
>
> notify url is `:8080/notify`

## Environments

| name| description |
| --- | --- |
|GIT_USER_NAME|user name for private git repository |
|GIT_PASSWORD|password for private git repository|
|NOTIFY_TOKEN|token for notify docserver|
|WEB_USER_NAME|basic auth username|
|WEB_PASSWORD|basic auth password|
|DEBUG|DEBUG=1 open debug mode|

> if `WEB_USER_NAME` is not empty basic auth will be open

## Config file
config file must named as `conf.yaml`

this file contains multi section.

the first section is `docserver` config,other sections are mkdoc project config ,those config's format as the same as `mkdoc` config.

- docserver section

| name| description |
| --- | --- |
|repo| repository to clone|
|branch| branch to clone|

- projects section

| name| description |
| --- | --- |
|id|path for doc page|

for example:
```yaml
repo: "https://github.com/TheWinds/mkdoc.git"
branch: develop
---
id: project_1
name: mkdoc example1
desc: this doc is auto generated by [mkdoc](https://github.com/TheWinds/mkdoc)
api_base_url: "http://localhost:8080"
mime:
  in:  form
  out: json
scanner:
  - gofunc
generator:
  - docsify
args:
  enable_go_mod: true
  path: "./src"
---
id: project_2
name: mkdoc example2
desc: this doc is auto generated by [mkdoc](https://github.com/TheWinds/mkdoc)
api_base_url: "http://localhost:8080"
mime:
  in:  form
  out: json
inject:
  - name: "token"
    desc: "jwt token"
    default: "hfjdjhkklashjkfsd.hjkfsdajhkfdsj.jknsfdksf"
    scope: header
scanner:
  - gofunc
  - docdef
generator:
  - markdown
  - insomnia
  - docsify
args:
  enable_go_mod: true
  path: "./src"
```  