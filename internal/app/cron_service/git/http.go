package git

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
)

// HTTPInfoRefs 处理 git smart http 协议的 info-refs 请求, 用于获取远程仓库的信息.
// @Summary
// @Description.
// @Tags Git
// @Accept json
// @Produce json
// @Param name path string true "仓库名称"
// @Param service query string true "服务类型"
// @Router /{name}/info/refs [get]
func HTTPInfoRefs(c *gin.Context) {
	repoName := c.Param("name")

	writer := c.Writer

	service := c.Query("service")
	writer.Header().Set("content-type", fmt.Sprintf("application/x-%s-advertisement", service))

	session, err := CreateSession(service, repoName)
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	if err := AdvertisedReferences(Context{Context: c, OverHTTP: true}, service, session, writer); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}
}

// HTTPUploadPack 处理 git smart http 协议的 upload-pack 请求, 用于客户端向服务端请求下载数据.
// @Summary
// @Description.
// @Tags Git
// @Accept json
// @Produce json
// @Param name path string true "仓库名称"
// @Router /{name}/git-upload-pack [post]
func HTTPUploadPack(c *gin.Context) {
	repoName := c.Param("name")

	reader := c.Request.Body
	writer := c.Writer

	service := "git-upload-pack"
	writer.Header().Set("content-type", fmt.Sprintf("application/x-%s-result", service))

	session, err := CreateSession(service, repoName)
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	if err := UploadPack(Context{Context: c, OverHTTP: true}, session.(transport.UploadPackSession), reader, writer); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}
}

// HTTPReceivePack 处理 git smart http 协议的 receive-pack 请求, 用于客户端向服务端请求上传数据.
// @Summary
// @Description.
// @Tags Git
// @Accept json
// @Produce json
// @Param name path string true "仓库名称"
// @Router /{name}/git-receive-pack [post]
func HTTPReceivePack(c *gin.Context) {
	repoName := c.Param("name")

	reader := c.Request.Body
	writer := c.Writer

	service := "git-receive-pack"
	writer.Header().Set("content-type", fmt.Sprintf("application/x-%s-result", service))

	session, err := CreateSession(service, repoName)
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	if err := ReceivePack(Context{Context: c, OverHTTP: true}, session.(transport.ReceivePackSession), reader, writer); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}
}
