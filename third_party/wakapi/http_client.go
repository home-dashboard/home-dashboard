package wakapi

import (
	"encoding/base64"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_http_client"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"net/http"
	"strings"
	"time"
)

// Wakapi 的鉴权请求头. 由 [configuration.Config.ServerMonitor.ThirdParty.Wakapi.ApiKey] 值的 base64 格式拼接而成.
var authorization string

// [configuration.Config.ServerMonitor.ThirdParty.Wakapi.ApiUrl] 的值.
var apiUrl string

var client *comfy_http_client.Client

// 初始化 Wakapi 的 http client.
func httpClientInitial() error {
	config := configuration.Config.ServerMonitor.ThirdParty.Wakapi

	// 拼接 Wakapi 的鉴权请求头.
	// 来源: https://wakapi.dev/swagger-ui/swagger-ui/index.html
	authorization = strings.Join([]string{"Basic", base64.StdEncoding.EncodeToString([]byte(config.ApiKey))}, " ")
	apiUrl = strings.Trim(config.ApiUrl, "/")

	if _client, err := comfy_http_client.New(apiUrl, http.Header{"authorization": []string{authorization}}, time.Second*10); err != nil {
		return err
	} else {
		client = _client
	}

	return nil
}
