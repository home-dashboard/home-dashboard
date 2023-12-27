package file_service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/file_service"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"path/filepath"
)

var localFileService *file_service.LocalFileService

func Get() (*file_service.LocalFileService, error) {
	if localFileService == nil {
		return nil, fmt.Errorf("file service is not initialized")
	}
	return localFileService, nil
}

// Serve 为给定目录中的静态文件提供静态文件服务.
func Serve(router *gin.RouterGroup) error {
	localFileService = file_service.NewLocalFileService(filepath.Join(utils.WorkspaceDir(), "temp_files"))

	router.StaticFS("/", localFileService.FileSystem)

	return nil
}
