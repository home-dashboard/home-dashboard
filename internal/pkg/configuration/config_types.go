package configuration

import (
	"encoding/json"
	"time"
)

type ServerMonitorConfiguration struct {
	// 服务监听端口, 默认为 8080
	Port uint `json:"port" toml:"port"`
	// 管理员账号信息
	Administrator ServerMonitorAdministratorConfiguration `json:"administrator" toml:"administrator"`
	// 开发模式配置
	Development ServerMonitorDevelopmentConfiguration `json:"development" toml:"development"`
	// 第三放服务配置
	ThirdParty ServerMonitorThirdPartyConfiguration `json:"thirdParty" toml:"thirdParty"`
	// 用于检查服务更新的配置
	Update ServerMonitorUpdateConfiguration `json:"update" toml:"update"`
}

type ServerMonitorAdministratorConfiguration struct {
	// 管理员用户名, 默认为 administrator
	Username string `json:"username" toml:"username"`
	// 管理员密码, 默认为 123456
	Password string `json:"password" toml:"password"`
}

type ServerMonitorDevelopmentConfiguration struct {
	// 是否为开发模式. 默认为 false
	// 开发模式下:
	// - overseer 将在同一进程中执行程序, 而不是子进程. 详见 [github.com/siaikin/home-dashboard/internal/pkg/overseer.Config.DebugMode]
	Enable bool `json:"enable" toml:"enable"`
	// 开发模式下的跨域配置
	Cors struct {
		// 允许跨域请求的源的列表, 该值将会被添加到 [Access-Control-Allow-Origin] 标头中.
		//
		// [Access-Control-Allow-Origin]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
		AllowOrigins []string `json:"allowOrigins" toml:"allowOrigins"`
	} `json:"cors" toml:"cors"`
}

// ServerMonitorThirdPartyConfiguration 第三方服务配置
type ServerMonitorThirdPartyConfiguration struct {
	Wakapi ServerMonitorThirdPartyWakapiConfiguration `json:"wakapi" toml:"wakapi"`
	GitHub ServerMonitorThirdPartyGitHubConfiguration `json:"github" toml:"github"`
}

// ServerMonitorThirdPartyWakapiConfiguration 第三方服务 Wakapi 的配置
type ServerMonitorThirdPartyWakapiConfiguration struct {
	// 是否从 Wakapi 收集数据
	// 默认为 false
	Enable bool `json:"enable" toml:"enable"`
	// Wakapi 用户的 api_key
	ApiKey string `json:"apiKey" toml:"apiKey"`
	// Wakapi 服务的地址
	ApiUrl string `json:"apiUrl" toml:"apiUrl"`
}

// ServerMonitorThirdPartyGitHubConfiguration 第三方服务 GitHub 的配置
type ServerMonitorThirdPartyGitHubConfiguration struct {
	// 是否从 GitHub 收集数据
	// 默认为 false
	Enable bool `json:"enable" toml:"enable"`
	// GitHub 访问令牌, 用于访问 GitHub API
	// See https://docs.github.com/zh/rest/overview/authenticating-to-the-rest-api?apiVersion=2022-11-28#%E4%BD%BF%E7%94%A8-personal-access-token-%E8%BF%9B%E8%A1%8C%E8%BA%AB%E4%BB%BD%E9%AA%8C%E8%AF%81
	PersonalAccessToken string `json:"personalAccessToken" toml:"personalAccessToken"`
}

// ServerMonitorUpdateConfiguration 用于检查服务更新的配置.
// 通过实现 [github.com/siaikin/home-dashboard/internal/pkg/overseer/fetcher.Fetcher] 接口来对接不同的更新检查服务.
type ServerMonitorUpdateConfiguration struct {
	// 是否启用自动更新
	// 默认为 false
	Enable bool `json:"enable" toml:"enable"`
	// 服务监听端口, 默认为 8081
	Port uint `json:"port" toml:"port"`
	// 更新检查的时间间隔, 单位为秒
	// 默认为 10 分钟(600 秒)
	FetchInterval time.Duration `json:"fetchInterval" toml:"fetchInterval"`
	// 更新检查的超时时间, 单位为秒. 详细说明见 [github.com/siaikin/home-dashboard/internal/pkg/overseer.Config.FetchTimeout]
	// 默认为 10 分钟(600 秒)
	FetchTimeout time.Duration `json:"fetchTimeout" toml:"fetchTimeout"`
	// 配置用于检查更新的 Fetcher.
	Fetchers ServerMonitorUpdateFetchersConfiguration `json:"fetchers" toml:"fetchers"`
}

// ServerMonitorUpdateFetchersConfiguration 用于检查服务更新的 Fetcher 的配置.
type ServerMonitorUpdateFetchersConfiguration struct {
	// GitHub Fetcher 的配置
	GitHub ServerMonitorUpdateFetcherGitHubConfiguration `json:"github" toml:"github"`
}

// ServerMonitorUpdateFetcherGitHubConfiguration 用于配置跟 GitHub 对接的 Fetcher.
// See [github.com/siaikin/home-dashboard/internal/pkg/overseer/fetcher.GitHubFetcher]
type ServerMonitorUpdateFetcherGitHubConfiguration struct {
	// GitHub 访问令牌, 用于访问 GitHub API. 未配置时, 将会使用 [ServerMonitorThirdPartyGitHubConfiguration.PersonalAccessToken] 的值.
	PersonalAccessToken string `json:"personalAccessToken" toml:"personalAccessToken"`
	// GitHub 仓库的拥有者
	Owner string `json:"owner" toml:"owner"`
	// GitHub 仓库的名称
	Repository string `json:"repository" toml:"repository"`
}

type Configuration struct {
	ServerMonitor ServerMonitorConfiguration `json:"serverMonitor" toml:"serverMonitor"`
	// 配置文件的修改时间, 值来自 [time.Time.UnixNano]
	ModificationTime int64 `json:"modificationTime"`
}

func (c Configuration) String() string {
	marshal, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(marshal)
}
