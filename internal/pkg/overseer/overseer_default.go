//go:build !windows

package overseer

import (
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
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
	if err := file.Chown(uid, gid); err != nil {
		return err
	}

	return nil
}

// sendTerminalSignal 向进程发送终止信号
func sendTerminalSignal(proc *os.Process) error {
	return proc.Signal(unix.SIGTERM)
}

func SetCommandSysProcAttr(cmd *exec.Cmd) {
	// no-op
}
