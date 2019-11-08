package docspace

import "log"

// APIScanner 实现该接口的扫描器可以从源码获取到API文档注解
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

var scanners map[string]APIScanner

// RegisterScanner to global scanners
func RegisterScanner(scanner APIScanner) {
	if scanners == nil {
		scanners = make(map[string]APIScanner)
	}
	scannerName := scanner.GetName()
	if scanners[scannerName] != nil {
		log.Fatalf("duplicate register scanner : %s", scannerName)
	}
	scanners[scannerName] = scanner
}

// GetScanners get all registered scanners
func GetScanners() map[string]APIScanner {
	return scanners
}
