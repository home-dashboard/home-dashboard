package utils

import (
	"os"
	"path/filepath"
)

// FileCreate 创建文件, 并自动创建缺少的目录. path 的最后一部分被认为是文件名.
func FileCreate(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}

	return os.Create(path)
}
