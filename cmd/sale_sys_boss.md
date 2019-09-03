# saleboss API
[TOC]
### 搜索销售线索
> note:saleStatus不为0时,必须填classAdviserId,此时其他搜索条件无效
- [type] query graphql
```
[path] /zhike/courseManage
```
- 参数
```json
{
    "nickname" : "thewinds",	# 用户昵称
    "childName" : "thewinds",	# 孩子姓名
    "classAdviserId" : "499379",	# 班主任ID
    "startTime" : 1567492259,	# 开课时间
    "endTime" : 1567492259,	# 结课时间
    "page" : 1,	# 页码
    "limit" : 0,	# 页码大小
    "saleStatus" : 0,	# 销售状态 0-无 1-跟进中 2-新分配 3-已转化
    "tagId" : 729906	# 标签ID
}
```
- 返回
```json
{
    "customerId" : 133274,	# 用户ID
    "nickname" : "thewinds",	# 用户昵称
    "avatar" : "",	# 用户头像
    "childName" : "thewinds",	# 孩子名称
    "grade" : "",	# 孩子年级
    "activeTime" : 1567492259,	# 课程激活时间
    "deadline" : 0,	# 课程结束时间
    "totalWork" : 0,	# 总作业数
    "teacherName" : "thewinds",	# 关联老师名称
    "courseId" : 984998,	# 课程ID
    "recentSaleTag" : "",	# 用户近期销售标签
    "lastFollowUpTime" : 1567492259	# 最后跟进时间
}
```
### 销售跟进统计
- [type] query graphql
```
[path] /zhike/courseManage
```
- 参数
```json
{
    "current" : 1,	# 当前页
    "pageSize" : 10,	# 页面大小
    "dayAgo" : 0	# 查询几天前的数据 默认0,昨天-1,前天-2,...
}
```
- 返回
```json
{
    "rows" : [1,2,3]
}
```
### 设置用户班主任
- [type] mutation graphql
```
[path] /zhike/courseManage
```
- 参数
```json
{
    "classAdviserId" : "902992",	# 
    "customerId" : 942792	# 
}
```
- 返回
```json
{
}
```
