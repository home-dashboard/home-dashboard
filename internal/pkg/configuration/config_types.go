package configuration

import (
	"encoding/json"
)

type ServerMonitorConfiguration struct {
	// 服务监听端口, 默认为 8080
	Port uint `json:"port" toml:"port"`
	// 管理员账号信息
	Administrator ServerMonitorAdministratorConfiguration `json:"administrator" toml:"administrator"`
	// 开发模式配置
	Development ServerMonitorDevelopmentConfiguration `json:"development" toml:"development"`
}

type ServerMonitorAdministratorConfiguration struct {
	// 管理员用户名, 默认为 administrator
	Username string `json:"username" toml:"username"`
	// 管理员密码, 默认为 123456
	Password string `json:"password" toml:"password"`
}

type ServerMonitorDevelopmentConfiguration struct {
	// 是否为开发模式
	// 默认为 false
	Enable bool `json:"enable" toml:"enable"`
	// 开发模式下的跨域配置
	Cors struct {
		// 允许跨域请求的源的列表, 该值将会被添加到 [Access-Control-Allow-Origin] 标头中.
		//
		// [Access-Control-Allow-Origin]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
		AllowOrigins []string `json:"allowOrigins" toml:"allowOrigins"`
	} `json:"cors" toml:"cors"`
}

type Configuration struct {
	ServerMonitor ServerMonitorConfiguration `json:"serverMonitor" toml:"serverMonitor"`
}

func (c Configuration) String() string {
	marshal, err := json.Marshal(c)
	if err != nil {
		return ""
	}

	return string(marshal)
}
