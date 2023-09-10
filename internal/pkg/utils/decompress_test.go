package utils_test

import (
	"archive/tar"
	"compress/gzip"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"os"
	"path/filepath"
	"testing"
)

func TestDecompressFile(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建一个测试文件进行解压缩
	testFile := filepath.Join(tempDir, "test.tar.gz")
	if err := createTestFile(testFile); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试解压缩文件
	destDir := filepath.Join(tempDir, "dest")
	if err := utils.DecompressFile(testFile, destDir); err != nil {
		t.Fatalf("解压缩文件失败: %v", err)
	}

	// 检查解压缩后的文件是否存在
	if _, err := os.Stat(filepath.Join(destDir, "test.txt")); os.IsNotExist(err) {
		t.Fatalf("解压缩后的文件不存在: %v", err)
	}

	// 测试解压缩目录
	if err := utils.DecompressFile(tempDir, destDir); err == nil {
		t.Fatalf("预期解压缩目录时出错，但得到了nil")
	}

	// 测试解压缩到相同的目录
	if err := utils.DecompressFile(testFile, tempDir); err == nil {
		t.Fatalf("预期解压缩到相同的目录时出错，但得到了nil")
	}

	// 测试解压缩没有扩展名的文件
	if err := utils.DecompressFile(filepath.Join(tempDir, "test"), destDir); err == nil {
		t.Fatalf("预期解压缩没有扩展名的文件时出错，但得到了nil")
	}

	// 测试解压缩一个目录
	dir := filepath.Join(tempDir, "dir")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := utils.DecompressFile(dir, destDir); err == nil {
		t.Fatalf("预期解压缩目录时出错，但得到了nil")
	}

	// 测试解压缩一个不存在的文件
	if err := utils.DecompressFile(filepath.Join(tempDir, "notexist.tar.gz"), destDir); err == nil {
		t.Fatalf("预期解压缩不存在的文件时出错，但得到了nil")
	}

	// 测试解压缩一个无效的 gzip 文件
	invalidGzFile := filepath.Join(tempDir, "invalid.tar.gz")
	if err := os.WriteFile(invalidGzFile, []byte("invalid gzip file"), 0644); err != nil {
		t.Fatalf("创建无效的 gzip 文件失败: %v", err)
	}
	if err := utils.DecompressFile(invalidGzFile, destDir); err == nil {
		t.Fatalf("预期解压缩无效的 gzip 文件时出错，但得到了nil")
	}
}

func createTestFile(filename string) error {
	// 创建一个测试文件进行压缩
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	gw := gzip.NewWriter(file)
	defer func(gw *gzip.Writer) {
		_ = gw.Close()
	}(gw)
	tw := tar.NewWriter(gw)
	defer func(tw *tar.Writer) {
		_ = tw.Close()
	}(tw)

	fileContents := []byte("测试文件内容")
	if err := tw.WriteHeader(&tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(fileContents)),
	}); err != nil {
		return err
	}
	if _, err := tw.Write(fileContents); err != nil {
		return err
	}

	return nil
}
