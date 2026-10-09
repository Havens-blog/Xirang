package config

import "os"

// writeFile 供测试写入配置文件（保持测试文件聚焦于用例本身）。
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
