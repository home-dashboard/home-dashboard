package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Move 移动文件, 支持跨设备移动. 如果是跨设备移动, 则会先复制文件, 然后删除源文件.
// 将 source 移动到 destination, 如果 destination 已存在, 则会覆盖.
func Move(source, destination string) error {
	err := os.Rename(source, destination)
	if err == nil {
		return nil
	}

	if !strings.Contains(err.Error(), "invalid cross-device link") && !strings.Contains(err.Error(), "The system cannot move the file to a different disk drive") {
		if exist, err := FileExist(destination); err != nil || !exist {
			return err
		}
		return nil
	}

	//fmt.Printf("move %s to %s cross device\n", source, destination)
	src, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source file error: %w", err)
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create destination file error: %w", err)
	}
	defer func() {
		_ = dst.Close()
	}()

	if _, err = io.Copy(dst, src); err != nil {
		_ = dst.Close()
		_ = os.Remove(destination)
		return fmt.Errorf("copy file error: %w", err)
	}

	sourceStat, err := os.Stat(source)
	if err != nil {
		_ = dst.Close()
		_ = os.Remove(destination)
		return fmt.Errorf("stat source file error: %w", err)
	}
	err = os.Chmod(destination, sourceStat.Mode())
	if err != nil {
		_ = dst.Close()
		_ = os.Remove(destination)
		return fmt.Errorf("chmod destination file error: %w", err)
	}

	_ = src.Close()
	if err := os.Remove(source); err != nil {
		return fmt.Errorf("remove source file error: %w", err)
	}

	return nil
}
