package web_submodules

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/cache_control"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"io/fs"
	"net/http"
	"strings"
)

var logger = comfy_log.New("[web_submodules]")

//go:embed all:home-dashboard-web-ui/dist
var homeDashboardWebUiAssets embed.FS

// 为什么不用 "/index.html" 作为 indexFilePath?
// 首先, 已知对于 "/" golang 会自动重定向到 "/index.html"(见 https://github.com/gin-gonic/gin/issues/2654#issuecomment-815823804 ).
// 当 indexFilePath 值为 "/index.html" 时, 如果找不到文件程序会在路径尾部增加 indexFilePath. 而 golang 内部会重定向到没有 "index.html" 后缀的相同路径.
// 这又会导致再次找不到文件, 从而陷入死循环.
var indexFilePath = "/"

func EmbedHomeDashboardWebUI(engine *gin.Engine) error {
	var extractedFs fs.FS
	var err error
	if extractedFs, err = fs.Sub(homeDashboardWebUiAssets, "home-dashboard-web-ui/dist/public"); err != nil {
		return err
	}

	httpExtractedFs := http.FS(extractedFs)
	engine.NoRoute(cache_control.CacheControlMiddleware("public, max-age=604800, immutable"), cache_control.ETagMiddleware(extractedFs), func(c *gin.Context) {
		filePath := strings.Trim(c.Request.URL.EscapedPath(), "/")

		if _, err = fs.ReadFile(extractedFs, filePath); err != nil {
			// cache_control.ETagMiddleware 会会为所有响应添加 Cache-Control 头.
			// 但是 index.html 需要每次请求通过 ETag 验证以确保能够获取最新的内容, 因此需要清空 Cache-Control 头.
			c.Header("Cache-Control", "")

			filePath = indexFilePath

			logger.Warn("filepath %s not found, redirect to %s\n", filePath, indexFilePath)
		}

		logger.Info("filepath %s\n", filePath)
		c.FileFromFS(filePath, httpExtractedFs)
	})

	logger.Info("home-dashboard-web-ui mount complete\n")
	return nil
}
