package comfy_log

import (
	"github.com/go-errors/errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var logger *Logger

var loggerMap = make(map[string]*Logger)

func init() {
	logger = newLogger("[comfy_log]")
}

// New 创建一个新的 Logger, prefix 参数将被添加到每一条日志的开头并以 Separator 作为与后面日志信息的分隔符.
func New(prefix string) *Logger {
	if loggerMap[prefix] != nil {
		return loggerMap[prefix]
	}

	loggerMap[prefix] = newLogger(prefix)
	return loggerMap[prefix]
}

var osStdout = os.Stdout
var osStderr = os.Stderr

// InterceptStandardOutput 拦截 os.Stdout, os.Stderr 的输出, 以便将日志输出到文件中.
func InterceptStandardOutput() error {
	var err error
	os.Stdout, err = standardCopyToFile(os.Stdout, false)
	if err != nil {
		return err
	}

	os.Stderr, err = standardCopyToFile(os.Stdout, true)
	if err != nil {
		return err
	}

	// 重新设置已创建的 Logger 的输出目标.
	for _, l := range loggerMap {
		l.stdout.SetOutput(os.Stdout)
		l.stderr.SetOutput(os.Stderr)

	}

	return nil
}

func standardCopyToFile(osStd *os.File, isStderr bool) (*os.File, error) {
	outR, outW, err := os.Pipe()
	if err != nil {
		return nil, errors.Errorf("intercept standard output failed, %w", err)
	}

	multiOut := io.MultiWriter(osStd, logFileWriter{writeToStdoutFile: !isStderr})
	go func() {
		if _, err := io.Copy(multiOut, outR); err != nil {
			logger.Fatal("stdout copy failed, %v\n", err)
		}
	}()

	return outW, nil
}

func newLogger(prefix string) *Logger {
	if !strings.HasSuffix(prefix, Separator) {
		prefix = strings.Join([]string{strconv.FormatInt(int64(os.Getpid()), 10), prefix, ""}, Separator)
	}

	return &Logger{
		stdout: log.New(os.Stdout, prefix, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
		stderr: log.New(os.Stderr, prefix, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
	}
}
