package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
)

// MD5 计算 MD5 值
func MD5(data *[]byte) (string, error) {
	hash := md5.New()
	if _, err := hash.Write(*data); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// MD5String 计算字符串的 MD5 值
func MD5String(str string) (string, error) {
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

// MD5FSFile 计算 path 文件的 MD5 值
func MD5FSFile(fsys fs.FS, path string) (string, error) {
	b, err := fs.ReadFile(fsys, path)
	if err != nil {
		return "", err
	}

	hash := md5.New()
	if _, err := hash.Write(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
