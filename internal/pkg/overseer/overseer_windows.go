//go:build windows

package overseer

import (
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"syscall"
)

const (
	TerminateSignal       = windows.SIGTERM
	InitialCompleteSignal = windows.SIGTERM
	TakeOverSignal        = windows.SIGTERM
)

var (
	uid = -1
	gid = -1
)

// 在 Windows 上, 需要以 .exe 为扩展名, 否则不执行任何操作.
func creteTempFile() (*os.File, error) {
	return utils.CreateTempFile("home-dashboard.exe")
}

// replaceExecutableFile 替换当前进程的二进制文件.
// 由于 Windows 上的文件锁定机制, 需要先将当前进程的二进制文件重命名为 .old, 然后再将新的二进制文件重命名为当前进程的二进制文件. 这像一个 hack, 但是目前没有更好的解决方案.
//
// 参考:
// [Download and replace running EXE]: https://www.codeproject.com/Questions/621666/Download-and-replace-running-EXE
func replaceExecutableFile(newFilePath, oldFilePath string) error {
	if err := utils.Move(oldFilePath, oldFilePath+".old"); err != nil {
		return err
	} else if err := utils.Move(newFilePath, oldFilePath); err != nil {
		return err
	}

	return nil
}

// todo 实现 windows 的 chown 方法
func chown(file *os.File, uid, gid int) error {
	return nil
}

func sendTerminalSignal(proc *os.Process) error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}

	p, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	r, _, err := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(proc.Pid))
	if r == 0 {
		return err
	}

	return nil
}

func SetCommandSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP,
	}
}
