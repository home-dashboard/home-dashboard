//go:build windows

package overseer

import (
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"syscall"
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
		return errors.New(err)
	} else if err := utils.Move(newFilePath, oldFilePath); err != nil {
		return errors.New(err)
	}

	return nil
}

// todo 实现 windows 的 chown 方法
func chown(file *os.File, uid, gid int) error {
	return nil
}

// sendTerminalSignal 向进程发送终止信号. 对于 Windows, 使用 [GenerateConsoleCtrlEvent function] 方法.
// 参考:
//   - [NOT SUPPORT send SIGINT to other process on windows?]
//   - [goreaman/proc_windows.go]
//
// 因为 Golang 并未实现 Windows 下的信号发送方法, 因此需要自行实现发送终止信号的方法.
// 实际上 Windows 中并不存在与 Linux 信号对等的机制, Windows 所谓的信号是通过控制台事件([GenerateConsoleCtrlEvent function])来实现的.
// ## 注意事项:
// 发送的控制台事件会被所有在同一进程组的进程接收到(包括发送事件的进程). 为了避免这种情况, 可以把新创建的进程放到一个新的进程组中. 这样在发送终止信号给所谓的"子进程"时, "父进程"就不会接收到该信号.
// 为了实现这一点, 需要在创建新进程时, 设置 SysProcAttr.CreationFlags 为 CREATE_NEW_PROCESS_GROUP. 这样新创建的进程就会被放到一个新的进程组中. 请参考 SetCommandSysProcAttr 方法.
//
// [NOT SUPPORT send SIGINT to other process on windows?]: https://github.com/golang/go/issues/28498
// [goreaman/proc_windows.go]: https://github.com/mattn/goreman/blob/ebb9736b7c7f7f3425280ab69e1f7989fb34eadc/proc_windows.go#L16
// [GenerateConsoleCtrlEvent function]: https://learn.microsoft.com/zh-cn/windows/console/generateconsolectrlevent
func sendTerminalSignal(proc *os.Process) error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}

	p, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return errors.New(err)
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
