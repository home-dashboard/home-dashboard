package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// MD5 计算字符串的 MD5 值
func MD5(str string) (string, error) {
	hash := md5.New()
	if _, err := io.WriteString(hash, str); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// MD5File 计算 path 文件的 MD5 值
func MD5File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
