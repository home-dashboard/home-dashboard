package comfy_log

import (
	"log"
	"os"
	"strings"
)

// New 创建一个新的 Logger, prefix 参数将被添加到每一条日志的开头并以 Separator 作为与后面日志信息的分隔符.
func New(prefix string) *Logger {
	if !strings.HasSuffix(prefix, Separator) {
		prefix = strings.Join([]string{prefix, ""}, Separator)
	}

	logger := &Logger{
		stdout: log.New(os.Stdout, prefix, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
		stderr: log.New(os.Stderr, prefix, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
	}

	return logger
}
