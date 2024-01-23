package wakapi

// [get-wakatimes-tats]: https://wakapi.dev/swagger-ui/swagger-ui/index.html#/wakatime/get-wakatimes-tats
// [get-wakatime-projects]: https://wakapi.dev/swagger-ui/swagger-ui/index.html#/wakatime/get-wakatime-projects
// [get-wakatime-summaries]: https://wakapi.dev/swagger-ui/swagger-ui/index.html#/wakatime/get-wakatime-summaries

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
	"net/url"
	"strings"
	"time"
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
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.EntityValidationError, err.Error()))
		return
	}

	timeRange := c.Param("range")
	// 从上下文中获取 userId , 作为请求路径的一部分.
	// 也可以使用 "current" 表示当前用户.
	userId, _ := c.Get("userId")

	if req, err := client.Get(strings.Join([]string{"compat/wakatime/v1/users/", userId.(string), "/stats/", timeRange}, "")); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
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
			_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
			return
		} else if str, err := client.ReadAsString(res); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
			return
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
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.EntityValidationError, err.Error()))
		return
	}

	req, err := client.Get("compat/wakatime/v1/users/current/projects")
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
	}
	res, err := client.Send(req)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
	}
	str, err := client.ReadAsString(res)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
	}

	c.String(http.StatusOK, str)
}

// summariesQueryParams 获取 Wakapi 摘要数据的查询参数. 详情见 [get-wakatime-summaries].
type summariesQueryParams struct {
	statsQueryParams
	Range string `form:"range"`
	Start int64  `form:"start"`
	End   int64  `form:"end"`
}

func GetSummaries(c *gin.Context) {
	var body summariesQueryParams
	var err error

	if err = c.ShouldBindQuery(&body); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.EntityValidationError, err.Error()))
		return
	}

	if req, err := client.Get("compat/wakatime/v1/users/current/summaries"); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
		return
	} else {
		startTime := time.UnixMilli(body.Start)
		endTime := time.UnixMilli(body.End)
		client.AppendQueryParams(req, url.Values{
			"range":            []string{body.Range},
			"start":            []string{startTime.Format("2006-01-02")},
			"end":              []string{endTime.Format("2006-01-02")},
			"project":          []string{body.Project},
			"language":         []string{body.Language},
			"editor":           []string{body.Editor},
			"operating_system": []string{body.OperatingSystem},
			"machine":          []string{body.Machine},
			"label":            []string{body.Label},
		})

		res, err := client.Send(req)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
			return
		}

		str, err := client.ReadAsString(res)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, err.Error()))
			return
		}

		c.String(http.StatusOK, str)
	}
}
