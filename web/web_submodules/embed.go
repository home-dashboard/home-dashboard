package web_submodules

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"io/fs"
	"net/http"
	"strings"
)

var logger = comfy_log.New("[web_submodules]")

//go:embed all:home-dashboard-web-ui/dist
var homeDashboardWebUiAssets embed.FS

// htm 而不是 html? https://github.com/gin-gonic/gin/issues/2654#issuecomment-815823804
var indexFilePath = "/"

func EmbedHomeDashboardWebUI(engine *gin.Engine) error {
	var extractedFs fs.FS
	var err error
	if extractedFs, err = fs.Sub(homeDashboardWebUiAssets, "home-dashboard-web-ui/dist/public"); err != nil {
		return err
	}

	httpExtractedFs := http.FS(extractedFs)
	engine.NoRoute(func(c *gin.Context) {
		filePath := strings.Trim(c.Request.URL.EscapedPath(), "/")

		if _, err = fs.ReadFile(extractedFs, filePath); err != nil {
			filePath = indexFilePath
			logger.Warn("filepath %s not found, redirect to %s\n", filePath, indexFilePath)
		}

		logger.Info("filepath %s\n", filePath)
		c.FileFromFS(filePath, httpExtractedFs)
	})

	logger.Info("home-dashboard-web-ui mount complete\n")
	return nil
}
