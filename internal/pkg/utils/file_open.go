package utils

import (
	"fmt"
	"os"
)

// FileOpenOnlyFile 打开文件, 如果文件不存在或者是目录则返回错误.
func FileOpenOnlyFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(file.Name())
	if err != nil {
		return nil, err
	}

	mode := info.Mode()
	// 文件类型是目录或者不是普通文件时返回错误.
	if mode.IsDir() {
		return nil, fmt.Errorf("file is dir: %s", file.Name())
	} else if !mode.IsRegular() {
		return nil, fmt.Errorf("file is not regular: %s", file.Name())
	} else {
		return file, nil
	}
}
