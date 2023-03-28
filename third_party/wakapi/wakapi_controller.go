package wakapi

// [get-wakatimes-tats]: https://wakapi.dev/swagger-ui/swagger-ui/index.html#/wakatime/get-wakatimes-tats
// [get-wakatime-projects]: https://wakapi.dev/swagger-ui/swagger-ui/index.html#/wakatime/get-wakatime-projects

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
	"net/url"
	"strings"
)

// statsQueryParams 获取 Wakapi 统计数据的查询参数. 详情见 [get-wakatimes-tats].
type statsQueryParams struct {
	Project         string `form:"project"`
	Language        string `form:"language"`
	Editor          string `form:"editor"`
	OperatingSystem string `form:"operatingSystem"`
	Machine         string `form:"machine"`
	Label           string `form:"label"`
}

// GetStats 获取 Wakapi 的统计数据. 详情见 [get-wakatimes-tats].
func GetStats(c *gin.Context) {
	var body statsQueryParams
	var err error

	if err = c.ShouldBindQuery(&body); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "bind query params failed, %w", err))
		panic(err)
	}

	_range := c.Param("range")
	// 从上下文中获取 userId , 作为请求路径的一部分.
	// 也可以使用 "current" 表示当前用户.
	userId, _ := c.Get("userId")

	if req, err := client.Get(strings.Join([]string{"/api/compat/wakatime/v1/users/", userId.(string), "/stats/", _range}, "")); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "create request failed, %w", err))
	} else {
		client.AppendQueryParams(req, url.Values{
			"project":          []string{body.Project},
			"language":         []string{body.Language},
			"editor":           []string{body.Editor},
			"operating_system": []string{body.OperatingSystem},
			"machine":          []string{body.Machine},
			"label":            []string{body.Label},
		})

		if res, err := client.Send(req); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "send request failed, %w", err))
		} else if str, err := client.ReadAsString(res); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "read response failed, %w", err))
		} else {
			c.String(http.StatusOK, str)
		}
	}
}

type projectListQueryParams struct {
	Project string `form:"project"`
}

// GetProjectList 获取 Wakapi 的项目列表. 详情见 [get-wakatime-projects].
func GetProjectList(c *gin.Context) {
	var body projectListQueryParams

	if err := c.ShouldBindQuery(&body); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "bind query params failed, %w", err))
		panic(err)
	}

	if req, err := client.Get("/api/compat/wakatime/v1/users/current/projects"); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "create request failed, %w", err))
	} else if res, err := client.Send(req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "send request failed, %w", err))
	} else if str, err := client.ReadAsString(res); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, "read response failed, %w", err))
	} else {
		c.String(http.StatusOK, str)
	}
}
