package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/nodejs"
	"golang.org/x/mod/semver"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

var logger = comfy_log.New("[cron_service nodejs]")

// ListNodejsVersion 列出所有 Node.js 版本.
// @Summary ListNodejsVersion
// @Description ListNodejsVersion
// @Tags ListNodejsVersion
// @Produce json
// @Success 200 {object} string
// @Router nodejs/version/list [get]
func ListNodejsVersion(c *gin.Context) {
	remote, err := url.Parse("https://nodejs.org")
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	proxy.Director = func(request *http.Request) {
		request.URL.Scheme = remote.Scheme
		request.URL.Host = remote.Host
		request.Host = remote.Host
		request.URL.Path = "/dist/index.json"
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// InstallNodejs 安装 Node.js.
// @Summary InstallNodejs
// @Description InstallNodejs
// @Tags InstallNodejs
// @Produce json
// @Param version path string true "Nodejs version, no prefix v"
// @Success 200 {object} string
// @Router nodejs/install [get]
func InstallNodejs(c *gin.Context) {
	version := strings.ToLower(c.Param("version"))

	if !semver.IsValid(version) {
		comfy_errors.ControllerUtils.RespondEntityValidationError(c, "invalid version")
		return
	}

	installer := nodejs.Installer{
		MirrorURL:     "https://nodejs.org",
		WorkDirectory: filepath.FromSlash("cron_service/nodejs"),
		OnProgress: func(written, total uint64) {
			logger.Info("download progress: %d/%d\n", written, total)
		},
	}

	// 如果已经安装了，先卸载
	if installer.IsInstalled() {
		if err := installer.Uninstall(); err != nil {
			comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
			return
		}
	}

	if err := installer.Install(version, runtime.GOOS, runtime.GOARCH); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}
}

// InstalledNodejsInfo 列出已安装的 Node.js 信息
// @Summary ListInstalledNodejs
// @Description ListInstalledNodejs
// @Tags ListInstalledNodejs
// @Produce json
// @Success 200 {object} string
// @Router nodejs/installed [get]
func InstalledNodejsInfo(c *gin.Context) {
	installer := nodejs.Installer{
		MirrorURL:     "https://nodejs.org",
		WorkDirectory: filepath.FromSlash("cron_service/nodejs"),
		OnProgress: func(written, total uint64) {
			logger.Info("download progress: %d/%d\n", written, total)
		},
	}

	if !installer.IsInstalled() {
		c.Status(http.StatusNotFound)
		return
	}

	info := installer.Info()

	c.JSON(http.StatusOK, info)
}
