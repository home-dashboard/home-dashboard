package utils

import (
	"os"
	"path/filepath"
)

// CreateTempFile 在 CreateTempDir 创建的临时目录中创建一个名为 filename 的临时文件.
func CreateTempFile(fileName string) (*os.File, error) {
	tempDir, err := CreateTempDir("")
	if err != nil {
		return nil, err
	}

	return os.Create(filepath.Join(tempDir, fileName))
}

// CreateTempDir 在 TempDir 返回的目录中创建一个临时文件夹.
// 临时文件夹的名称根据 pattern 生成. 生成规则同 os.MkdirTemp
func CreateTempDir(pattern string) (string, error) {
	dir := TempDir()

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return os.MkdirTemp(dir, pattern)
}

const temporaryDir = "home-dashboard_temp"

// TempDir 返回自定义的临时文件夹路径
func TempDir() string {
	return filepath.Join(os.TempDir(), temporaryDir)
}
