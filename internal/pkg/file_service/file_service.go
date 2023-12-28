package file_service

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
}

type FileService struct {
	FileSystem ServeFileSystem
	// Serve 返回一个中间件处理程序, 该处理程序在给定目录中提供静态文件.
	Serve func(urlPrefix string, fs ServeFileSystem) gin.HandlerFunc
	Save  func(dest string, file os.File) error
}
