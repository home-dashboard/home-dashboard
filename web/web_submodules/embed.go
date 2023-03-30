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
	var unwrappedFs, appFs, assetsFs fs.FS
	var err error
	if unwrappedFs, err = fs.Sub(homeDashboardWebUiAssets, "home-dashboard-web-ui/build"); err != nil {
		return err
	}
	// appFs 用于输出 home-dashboard-web-ui 的 css/js 文件
	if appFs, err = fs.Sub(unwrappedFs, "_app"); err != nil {
		return err
	}
	// assetsFs 用于输出 home-dashboard-web-ui 的静态资源文件
	if assetsFs, err = fs.Sub(unwrappedFs, "_assets"); err != nil {
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

	engine.StaticFS("/_app", http.FS(appFs))
	engine.StaticFS("/_assets", http.FS(assetsFs))

	logger.Info("home-dashboard-web-ui mount complete\n")
	return nil
}
