package arguments

import (
	"flag"
)

var (
	ServerPort = flag.Int("port", -1, "serve port")
	MockMode   = flag.Bool("mock", false, "enable mock mode")
)

func init() {
	flag.Parse()
}
