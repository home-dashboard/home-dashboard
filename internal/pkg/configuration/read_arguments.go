package configuration

import (
	"flag"
)

var (
	serverPort = flag.Uint("port", 0, "serve port")
	mockMode   = flag.Bool("mock", false, "enable mock mode")
)

var argumentsConfig *Configuration = parseArguments()

func parseArguments() *Configuration {
	flag.Parse()

	return &Configuration{
		ServerMonitor: ServerMonitorConfiguration{
			Port: *serverPort,
			Mock: *mockMode,
		}}
}
