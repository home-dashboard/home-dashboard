package wakapi

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"net/http"
	"net/url"
)

var logger = comfy_log.New("[THIRD-PARTY wakapi]")

// Use 启用 Wakapi 服务.
func Use(router *gin.RouterGroup) error {
	wakapiConfig := configuration.Config.ServerMonitor.ThirdParty.Wakapi

	// 校验参数
	if !wakapiConfig.Enable {
		return fmt.Errorf("third party service are used, but not enabled in configuration file")
	} else if parsedUrl, err := url.ParseRequestURI(wakapiConfig.ApiUrl); err != nil || len(parsedUrl.Scheme) <= 0 {
		return fmt.Errorf("api url [%s] cannot parsed. %s\n", wakapiConfig.ApiUrl, err)
	} else if len(wakapiConfig.ApiKey) <= 0 {
		return fmt.Errorf("api key is empty")
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
}

// 将 Wakapi appKey 对应的 userId 附加到请求上下文中.
func getUserIdMiddleware() gin.HandlerFunc {
	// 缓存 userId.
	var userId string

	getUserId := func() (string, error) {
		if len(userId) > 0 { // 如果已经缓存了 userId, 则直接返回.
			return userId, nil
		} else if req, err := client.Get("summary"); err != nil {
			return "", err
		} else {
			client.AppendQueryParams(req, url.Values{"interval": []string{"today"}})

			if res, err := client.Send(req); err != nil {
				return "", err
			} else if str, err := client.ReadAsString(res); err != nil {
				return "", err
			} else {
				var result map[string]any

				if err := json.Unmarshal([]byte(str), &result); err != nil {
					return "", err
				} else {
					return result["user_id"].(string), nil
				}
			}
		}
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
