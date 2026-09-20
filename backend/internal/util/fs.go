package util

import "os"

// writeFile 写文件（报告 PDF 落盘）。
func WriteFile(path string, content []byte) error {
	dir := path[:lastSlash(path)]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func lastSlash(path string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return i
		}
	}
	return len(path)
}
