package configuration

import (
	"encoding/json"
)

type ServerMonitorConfiguration struct {
	// 服务监听端口, 默认为 8080
	Port uint `json:"port" toml:"port"`
	// 是否为调试模式, 默认不开启调试模式
	Mock bool `json:"mock" toml:"mock"`
	// 管理员账号信息
	Administrator ServerMonitorAdministratorConfiguration `json:"administrator" toml:"administrator"`
}

type ServerMonitorAdministratorConfiguration struct {
	// 管理员用户名, 默认为 administrator
	Username string `json:"username" toml:"username"`
	// 管理员密码, 默认为 123456
	Password string `json:"password" toml:"password"`
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
