package file_service

import (
	"fmt"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type LocalFileService struct {
	FileSystem http.FileSystem
	Root       string
}

func NewLocalFileService(root string) *LocalFileService {
	return &LocalFileService{Root: root, FileSystem: http.Dir(root)}
}

// Save 保存文件到指定目录. dest 是相对于 Root 的路径.
func (l *LocalFileService) Save(dest string, file io.Reader) error {
	fullPath := filepath.Join(l.Root, dest)
	if exist, err := utils.FileExist(fullPath); err != nil {
		return err
	} else if exist {
		return fmt.Errorf("file %s already exist", fullPath)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return err
	} else if newFile, err := os.Create(fullPath); err != nil {
		return err
	} else if _, err := io.Copy(newFile, file); err != nil {
		return err
	}

	return nil
}
