package comfy_log

import (
	"context"
	"fmt"
	"github.com/jinzhu/now"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var logger = &Logger{
	stdout: log.New(os.Stdout, "[comfy_log]"+Separator, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
	stderr: log.New(os.Stderr, "[comfy_log]"+Separator, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
}

var stdoutFile *os.File
var stderrFile *os.File

// New 创建一个新的 Logger, prefix 参数将被添加到每一条日志的开头并以 Separator 作为与后面日志信息的分隔符.
func New(prefix string) *Logger {
	if !strings.HasSuffix(prefix, Separator) {
		prefix = strings.Join([]string{prefix, ""}, Separator)
	}

	stdout := io.MultiWriter(os.Stdout, logFileWriter{writeToStdoutFile: true})
	stderr := io.MultiWriter(os.Stderr, logFileWriter{writeToStdoutFile: false})

	_logger := &Logger{
		stdout: log.New(stdout, prefix, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
		stderr: log.New(stderr, prefix, log.LstdFlags|log.LUTC|log.Lshortfile|log.Lmsgprefix),
	}

	return _logger
}

func init() {
	// todo: 补充测试用例.
	// 1. 到达指定时间时创建新的日志文件并将其后的输出日志写入新的日志文件.

	// 每天 00:00:00 时替换日志文件.
	go func(context context.Context) {
		timer := time.NewTimer(getReplaceLogFileDuration())

		// 初始化时立即获取今天的日志文件
		replaceLogFile()

		for {
			select {
			case <-context.Done():
				timer.Stop()
				return
			case <-timer.C:
				// 替换日志文件.
				replaceLogFile()

				// 重置定时器.
				timer.Reset(getReplaceLogFileDuration())
			}
		}
	}(context.Background())
}

// getReplaceLogFileDuration 获取替换日志文件的时间间隔.
// 该时间间隔为当前时间到第二天 00:00:00 的时间间隔.
// 计算第二天开始而不是第一天的结束, 是为了避免在第一天的结束和第二天的开始之间产生日志文件.
func getReplaceLogFileDuration() time.Duration {
	return now.With(time.Now().AddDate(0, 0, 1)).BeginningOfDay().Sub(time.Now())
}

func replaceLogFile() {
	// 关闭旧的日志文件.
	if stdoutFile != nil {
		_ = stdoutFile.Close()
	}
	if stderrFile != nil {
		_ = stderrFile.Close()
	}

	// 打开新的日志文件.
	stdoutFile = openOrCreateLogFile("")
	stderrFile = openOrCreateLogFile("error")
}

// openOrCreateLogFile 打开日志文件, 如果日志文件不存在则创建.
func openOrCreateLogFile(suffix string) *os.File {
	dirPath := filepath.Join(utils.WorkspaceDir(), "logs")

	// 检查日志目录是否存在, 不存在则创建.
	exist, err := utils.FileExist(dirPath)
	if err != nil {
		logger.Fatal("check log dir exist failed, %s\n", err)
	} else if !exist {
		logger.Info("detect log dir not exist, try to create it\n")
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			logger.Fatal("create log dir failed, %s\n", err)
		}
	}

	logDir, err := os.Open(dirPath)
	if err != nil {
		logger.Fatal("open log dir failed, %s\n", err)
	}

	// 检查日志目录是否为目录.
	if fileInfo, err := logDir.Stat(); err != nil {
		logger.Fatal("get log dir stat failed, %s\n", err)
	} else if !fileInfo.IsDir() {
		logger.Fatal("log dir(%s) is not a directory\n", dirPath)
	}

	// 打开日志文件, 如果不存在则创建.
	filePath := filepath.Join(dirPath, generateLogFileName(time.Now(), suffix))
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		logger.Fatal("open log file failed, %s\n", err)
	}

	return logFile
}

// generateLogFileName 根据给的 date 生成日志文件名. 如果 suffix 不为空, 则在日志文件名后添加 suffix.
// 文件名格式为 "YYYYMMDD_dd-hh-mm[_suffix]?.log", 其中 "YYYYMMDD_dd-hh-mm" 表示日期, "suffix" 表示文件名后缀.
func generateLogFileName(date time.Time, suffix string) string {
	path := fmt.Sprintf("%04d%02d%02d_%02d-%02d-%02d", date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second())
	if len(suffix) > 0 {
		path = path + "_" + suffix
	}

	return path + ".log"
}
