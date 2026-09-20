package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SaveUploadFile 保存上传文件到 UPLOAD_DIR，返回可访问的相对 URL。
func SaveUploadFile(dir, field string, content []byte, filename string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir upload dir: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".png"
	}
	name := fmt.Sprintf("%s_%d%s", field, time.Now().UnixNano(), ext)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("write upload file: %w", err)
	}
	return "/uploads/" + name, nil
}

// ReadAll 读取 io.Reader 全部内容。
func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
