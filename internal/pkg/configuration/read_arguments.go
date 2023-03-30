package configuration

import (
	"flag"
)

var (
	serverPort = flag.Uint("port", 0, "serve port")
	devMode    = flag.Bool("development", false, "enable development mode")
)

var argumentsConfig *Configuration

func parseArguments() *Configuration {
	if argumentsConfig == nil {
		argumentsConfig = &Configuration{
			ServerMonitor: ServerMonitorConfiguration{
				Port:        *serverPort,
				Development: ServerMonitorDevelopmentConfiguration{Enable: *devMode},
			}}
	}

	return argumentsConfig
}
