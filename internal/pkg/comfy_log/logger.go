package comfy_log

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger struct {
	// 通过该 log.Logger 实例输出的日志将被输出到 os.Stdout
	stdout *log.Logger
	// 通过该 log.Logger 实例输出的日志将被输出到 os.Stderr
	stderr *log.Logger
}

const (
	InfoPrefix  = "[INFO]"
	warnPrefix  = "[WARN]"
	ErrorPrefix = "[ERROR]"
	FatalPrefix = "[FATAL]"
	PanicPrefix = "[PANIC]"
	// Separator 等级前缀与消息之间的分隔符
	Separator = " "
)

// Info 用于记录流程在执行时的日志(如服务启停, 配置信息等). 在错误发生时这些信息有助于确定错误发生的原因.
func (l *Logger) Info(format string, v ...any) {
	_ = l.stdout.Output(2, fmt.Sprintf(joinPrefix(InfoPrefix, format, Separator), v...))
}

// Warn 用于记录流程出现异常但可能会在之后一段时间恢复(也可能不会恢复, 一般来说这将导致 Error 日志被记录)的日志. 如重试操作, 无法识别的非必要参数等.
func (l *Logger) Warn(format string, v ...any) {
	_ = l.stdout.Output(2, fmt.Sprintf(joinPrefix(warnPrefix, format, Separator), v...))
}

// Error 用于记录导致流程中断的日志, 这些错误日志会提示用户需要进行某些操作以使得程序能够正常运行.
func (l *Logger) Error(format string, v ...any) {
	_ = l.stderr.Output(2, fmt.Sprintf(joinPrefix(ErrorPrefix, format, Separator), v...))
}

// Fatal 在输出带有 FatalPrefix 前缀的日志后调用 os.Exit(1)
func (l *Logger) Fatal(format string, v ...any) {
	_ = l.stderr.Output(2, fmt.Sprintf(joinPrefix(FatalPrefix, format, Separator), v...))
	os.Exit(1)
}

// Panic 在输出带有 PanicPrefix 前缀的日志后调用 panic()
func (l *Logger) Panic(format string, v ...any) {
	s := fmt.Sprintf(joinPrefix(PanicPrefix, format, Separator), v...)
	_ = l.stderr.Output(2, s)
	panic(s)
}

func joinPrefix(prefix string, format string, Separator string) string {
	return strings.Join([]string{prefix, format}, Separator)
}
