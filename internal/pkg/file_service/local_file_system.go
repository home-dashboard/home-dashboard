package file_service

import (
	"github.com/go-errors/errors"
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

	exist, err := utils.FileExist(fullPath)
	if err != nil {
		return errors.New(err)
	}
	if exist {
		return errors.Errorf("file %s already exist", fullPath)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return errors.New(err)
	}

	newFile, err := os.Create(fullPath)
	if err != nil {
		return errors.New(err)
	}

	if _, err := io.Copy(newFile, file); err != nil {
		return err
	}

	return nil
}
