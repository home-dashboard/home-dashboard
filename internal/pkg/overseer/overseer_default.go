//go:build !windows

package overseer

import (
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/sys/unix"
	"os"
)

const (
	// TerminateSignal 通知 worker 退出
	TerminateSignal = unix.SIGTERM
	// InitialCompleteSignal worker 通知 manager 初始化完成
	InitialCompleteSignal = unix.SIGUSR1
	// TakeOverSignal manager 通知 worker 可以接管网络监听
	TakeOverSignal = unix.SIGUSR1
)

var (
	uid = unix.Getuid()
	gid = unix.Getgid()
)

func creteTempFile() (*os.File, error) {
	return os.CreateTemp("home-dashboard_temp", "home-dashboard_*")
}

// replaceExecutableFile 替换当前进程的二进制文件
func replaceExecutableFile(newFilePath, oldFilePath string) error {
	return utils.Move(newFilePath, oldFilePath)
}

func chown(file *os.File, uid, gid int) error {
	// 继承 uid, gid
	if err := tempFile.Chown(uid, gid); err != nil {
		return err
	}
}

func sendTerminalSignal(proc *os.Process) error {
	return proc.Signal(m.Config.TerminateSignal)
}

func SetCommandSysProcAttr(cmd *exec.Cmd) {
	// no-op
}
