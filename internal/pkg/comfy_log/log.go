package comfy_log

import (
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var logger = New("[comfy_log]")

// New 创建一个新的 Logger, prefix 参数将被添加到每一条日志的开头并以 Separator 作为与后面日志信息的分隔符.
func New(prefix string) *Logger {
	if !strings.HasSuffix(prefix, Separator) {
		prefix = strings.Join([]string{strconv.FormatInt(int64(os.Getpid()), 10), prefix, ""}, Separator)
	}

	stdout := io.MultiWriter(os.Stdout, logFileWriter{writeToStdoutFile: true})
	stderr := io.MultiWriter(os.Stderr, logFileWriter{writeToStdoutFile: false})

	return &Logger{
		stdout: log.New(stdout, prefix, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
		stderr: log.New(stderr, prefix, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
	}
}
