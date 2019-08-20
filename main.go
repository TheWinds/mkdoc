package main

func main() {
	declare:=`@apidoc name name1`
	DocAnnotation(declare).ParseToAPI()
	//return
	scanGraphQLAPIDocInfo("corego/service/boss/schemas")
}