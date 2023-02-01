package configuration

import (
	"encoding/json"
)

type ServerMonitorConfiguration struct {
	Port uint `json:"port" toml:"port"`
	Mock bool `json:"mock" toml:"mock"`
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
