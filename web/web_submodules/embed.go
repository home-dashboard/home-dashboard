package web_submodules

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"io/fs"
	"net/http"
)

var logger = comfy_log.New("[web_submodules]")

//go:embed all:home-dashboard-web-ui/build
var homeDashboardWebUiAssets embed.FS

func EmbedHomeDashboardWebUI(engine *gin.Engine) error {
	var unwrappedFs, assetsFs fs.FS
	var err error
	if unwrappedFs, err = fs.Sub(homeDashboardWebUiAssets, "home-dashboard-web-ui/build"); err != nil {
		return err
	}
	if assetsFs, err = fs.Sub(unwrappedFs, "_app"); err != nil {
		return err
	}

	engine.StaticFileFS("/favicon.ico", "/favicon.png", http.FS(unwrappedFs))

	engine.NoRoute(func(c *gin.Context) {
		c.Writer.WriteHeader(200)

		raw, err := fs.ReadFile(unwrappedFs, "index.html")
		if err != nil {
			logger.Warn("read index.html failed, %s\n", err)
			return
		}

		if _, err := c.Writer.Write(raw); err != nil {
			logger.Warn("index.html write failed, %s\n", err)
			return
		}
	})

	engine.StaticFS("/_app", http.FS(assetsFs))

	logger.Info("assets load complete\n")
	return nil
}
