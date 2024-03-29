package wakapi

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"net/http"
	"net/url"
)

var logger = comfy_log.New("[THIRD-PARTY wakapi]")

// Use 启用 Wakapi 服务.
func Use(router *gin.RouterGroup) error {
	wakapiConfig := configuration.Get().ServerMonitor.ThirdParty.Wakapi

	// 校验参数
	if !wakapiConfig.Enable {
		return comfy_errors.NewResponseError(comfy_errors.UnknownError, "third party service are not enabled in configuration file")
	} else if parsedUrl, err := url.ParseRequestURI(wakapiConfig.ApiUrl); err != nil || len(parsedUrl.Scheme) <= 0 {
		return comfy_errors.NewResponseError(comfy_errors.UnknownError, "api url [%s] cannot parsed. %s\n", wakapiConfig.ApiUrl, err)
	} else if len(wakapiConfig.ApiKey) <= 0 {
		return comfy_errors.NewResponseError(comfy_errors.UnknownError, "api key is empty")
	}

	logger.Info("enabled with apiKey: \"%s\", apiUrl: \"%s\"\n", wakapiConfig.ApiKey, wakapiConfig.ApiUrl)

	if err := httpClientInitial(); err != nil {
		return err
	}

	initial(router)

	return nil
}

func initial(router *gin.RouterGroup) {
	router.Use(getUserIdMiddleware())

	router.GET("stats/:range", GetStats)
	router.GET("projects", GetProjectList)
	router.GET("summaries", GetSummaries)
}

// 将 Wakapi appKey 对应的 userId 附加到请求上下文中.
func getUserIdMiddleware() gin.HandlerFunc {
	// 缓存 userId.
	var userId string

	getUserId := func() (string, error) {
		if len(userId) > 0 { // 如果已经缓存了 userId, 则直接返回.
			return userId, nil
		}

		req, err := client.Get("summary")
		if err != nil {
			return "", err
		}

		client.AppendQueryParams(req, url.Values{"interval": []string{"today"}})

		res, err := client.Send(req)
		if err != nil {
			return "", err
		}

		str, err := client.ReadAsString(res)
		if err != nil {
			return "", err
		}

		var result map[string]any

		if err := json.Unmarshal([]byte(str), &result); err != nil {
			return "", errors.New(err)
		}

		return result["user_id"].(string), nil
	}

	return func(c *gin.Context) {
		if userId, err := getUserId(); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "get user id failed: %w", err))
		} else {
			// 将 userId 附加到请求上下文中.
			c.Set("userId", userId)
		}
	}
}
