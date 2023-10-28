package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrNotArchive = errors.New("not an archive")
)

// DecompressFile 解压文件到目标目录
func DecompressFile(source string, destinationDir string) error {
	// 校验参数并准备解压
	if err := validateAndPrepareDecompressArgs(source, destinationDir); err != nil {
		return err
	}

	// 打开源文件
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	extractFileName := source

	for {
		// 根据文件扩展名执行不同的解压操作
		fileOrDirPath, err := extractByExt(extractFileName, destinationDir)
		if err != nil {
			if errors.Is(err, ErrNotArchive) {
				break
			} else {
				return err
			}
		}
		extractFileName = fileOrDirPath
	}

	return nil
}

// extractByExt 根据文件扩展名执行不同的解压操作.
// 返回值: 解压后的文件或目录路径, 错误信息
func extractByExt(source string, destinationDir string) (string, error) {
	// 获取文件扩展名
	ext := filepath.Ext(source)
	// 设置解压后的文件名
	extractFileName := source[0 : len(source)-len(ext)]

	dest := destinationDir

	// 根据文件扩展名执行不同的解压操作
	switch ext {
	case ".gz": // 解压 gz 文件
		if _dest, err := extractGz(source, filepath.Join(destinationDir, filepath.Base(extractFileName))); err != nil {
			return "", err
		} else {
			dest = _dest
		}
	case ".tar": // 解压 tar 文件
		if err := extractTar(source, destinationDir); err != nil {
			return "", err
		}
	case ".zip": // 解压zip文件
		if err := extractZip(source, destinationDir); err != nil {
			return "", err
		}
	default:
		return "", ErrNotArchive
	}

	return dest, nil
}

func extractGz(source string, destination string) (string, error) {
	// 打开源文件
	file, err := os.Open(source)
	if err != nil {
		return "", err
	}
	stat, err := file.Stat()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	decompressedStream, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(destination), stat.Mode()); err != nil {
		return "", err
	}

	// 创建目标文件
	targetFile, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, stat.Mode())
	if err != nil {
		return "", err
	}
	defer func() {
		_ = targetFile.Close()
	}()

	// 将解压缩的文件内容复制到目标文件
	if _, err = io.Copy(targetFile, decompressedStream); err != nil {
		return "", err
	}

	return destination, nil
}

func extractTar(source string, destinationDir string) error {
	// 打开源文件
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	// 创建一个新的tar reader
	trr := tar.NewReader(file)

	// 从归档中提取每个文件
	for {
		// 从tar归档中获取下一个文件头
		hdr, err := trr.Next()
		if err == io.EOF {
			// 归档结束
			break
		}
		if err != nil {
			// 读取归档时出错
			return err
		}

		// 创建目标文件路径
		path := filepath.Join(destinationDir, hdr.Name)

		// 如果文件是目录，则创建它
		if hdr.FileInfo().IsDir() {
			if err := os.MkdirAll(path, hdr.FileInfo().Mode()); err != nil {
				return err
			}
		} else {
			// 提取文件
			if err := extractTarPart(trr, hdr, path); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractTarPart(trr *tar.Reader, hdr *tar.Header, destinationDir string) error {
	// 为文件创建目录结构
	if err := os.MkdirAll(filepath.Dir(destinationDir), hdr.FileInfo().Mode()); err != nil {
		return err
	}

	// 创建目标文件
	file, err := os.OpenFile(destinationDir, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
	if err != nil {
		return err
	}

	// 在函数返回时关闭文件
	defer func() {
		_ = file.Close()
	}()

	// 将文件内容从tar归档复制到目标文件
	if _, err = io.Copy(file, trr); err != nil {
		return err
	}

	return nil
}

func extractZip(source string, destinationDir string) error {
	// 打开zip文件
	r, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer func() {
		_ = r.Close()
	}()

	// 遍历zip文件中的所有文件
	for _, file := range r.File {
		if err := extractZipPart(file, destinationDir); err != nil {
			return err
		}
	}

	return nil
}

func extractZipPart(file *zip.File, destinationDir string) error {
	// 获取 zip.File 的完整路径
	path := filepath.Join(destinationDir, file.Name)

	// 处理目录类型
	if file.FileInfo().IsDir() {
		return os.MkdirAll(path, file.Mode())
	}

	// 打开zip文件中的文件
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = rc.Close()
	}()

	// 创建目标文件所在的目录
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	// 创建目标文件
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()

	// 将zip文件中的文件内容复制到目标文件
	if _, err = io.Copy(dst, rc); err != nil {
		return err
	}

	return nil
}

func validateAndPrepareDecompressArgs(source string, destinationDir string) error {
	// 判断源文件和目标目录是否相同
	if filepath.Dir(source) == destinationDir {
		return fmt.Errorf("source and destination are the same")
	}

	stat, err := os.Stat(source)
	// 判断源文件是否存在
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source file %s does not exist", source)
		} else {
			return err
		}
	}
	if stat.IsDir() {
		return fmt.Errorf("source file is a directory")
	}

	// 校验文件扩展名
	ext := filepath.Ext(source)
	if len(ext) <= 0 {
		return fmt.Errorf("file %s has no extension", source)
	}

	// 创建目标目录
	if err := os.MkdirAll(destinationDir, os.ModePerm); err != nil {
		return err
	}

	return nil
}
