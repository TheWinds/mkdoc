package docspace

// API扫描器
// 实现该接口的扫描器可以从源码获取到API文档注解
type APIScanner interface {
	// 扫描API注解
	ScanAnnotations(pkg string) ([]DocAnnotation, error)
	// 获取名称
	GetName() string
	// 设置配置
	SetConfig(map[string]interface{})
	// 获取扫描器相关帮助
	GetHelp() string
}
