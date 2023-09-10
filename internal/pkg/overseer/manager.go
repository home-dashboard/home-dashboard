package overseer

import (
	"fmt"
	"github.com/siaikin/home-dashboard/internal/pkg/overseer/fetcher"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/net/context"
	"io"
	"os"
	"os/exec"
	"time"
)

type managerStatus int

const (
	// managerStatusRunning manager 正在运行
	managerStatusRunning managerStatus = iota
	// managerStatusRestarting manager 正在重启
	managerStatusRestarting
	managerStatusDestroyed
)

// manager 在当前进程运行, 负责管理 worker 的生命周期.
type manager struct {
	Config Config
	Worker workerProxy
	Status managerStatus
	// BinPath 用于标识当前进程的二进制文件的路径.
	BinPath string
	// BinHash 用于标识当前进程的二进制文件的 hash 值.
	BinHash string
	// BinPermissions 用于标识当前进程的二进制文件的权限.
	BinPermissions os.FileMode
}

// Initial 将初始化 manager, 并开始检查更新.
func (m *manager) Initial(ctx context.Context) error {
	logger.Info("initializing manager")

	// 设置初始状态值
	m.Status = managerStatusRunning

	if err := m.setBinaryInfo(); err != nil {
		return err
	}

	// 启动时运行一次 worker
	logger.Info("running worker for the first time")
	if err := m.runWorker(ctx); err != nil {
		return err
	}

	// 检查版本更新
	go m.checkNewVersions(ctx)

	logger.Info("manager initialized")

	return nil
}

// Destroy 会销毁 manager, 并终止 worker.
func (m *manager) Destroy(ctx context.Context) {
	// 终止 worker
	_ = m.terminateWorker(ctx)

	// 标记状态为已销毁
	m.Status = managerStatusDestroyed
}

func (m *manager) setBinaryInfo() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}

	// 获取当前进程的二进制文件的路径
	m.BinPath = path

	// 获取当前进程的二进制文件的 hash 值
	if hash, err := utils.MD5File(path); err != nil {
		return err
	} else {
		m.BinHash = hash
	}

	// 获取当前进程的二进制文件的权限
	if info, err := os.Stat(path); err != nil {
		return err
	} else {
		m.BinPermissions = info.Mode()
	}

	return nil
}

func (m *manager) checkNewVersions(ctx context.Context) {
	timer := time.NewTimer(m.Config.FetchInterval)

	for _, item := range m.Config.Fetchers {
		if err := item.Init(); err != nil {
			logger.Error("failed to initialize fetcher: %v", err)
		}
	}

	for {
		select {
		case <-timer.C:
			for _, item := range m.Config.Fetchers {
				timeoutCtx, _ := context.WithTimeout(ctx, m.Config.FetchTimeout)
				// 检查是否有新版本可用
				if err := m.checkNewVersion(timeoutCtx, item); err != nil {
					logger.Error("failed to check new version: %v", err)
				}
			}
			timer.Reset(m.Config.FetchInterval)
		case <-ctx.Done():
			return
		}
	}
}

// runWorker 启动一个子进程. 当子进程成功启动并交换信令后返回.
func (m *manager) runWorker(ctx context.Context) error {
	path, err := os.Executable()
	if err != nil {
		return err
	}

	extendedEnv := os.Environ()
	hash, err := utils.MD5File(path)
	if err != nil {
		return err
	}

	// hash 取最后 8 位, 用于子进程启动时校验二进制文件.
	extendedEnv = append(extendedEnv, envIsWorkerProcess+"="+"true", envIsManagerProcess+"="+"false", envShortBinHash+"="+hash[len(hash)-8:])

	cmd := exec.Command(path)
	cmd.Env = extendedEnv
	// 继承当前进程的命令行参数
	cmd.Args = os.Args[1:]
	// 继承当前进程的 stdin, stdout, stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置系统调用参数
	SetCommandSysProcAttr(cmd)

	// 启动子进程
	if err := cmd.Start(); err != nil {
		return err
	}
	logger.Info("worker started")

	m.Worker = workerProxy{
		Cmd: cmd,
	}

	return nil
}

// terminateWorker 会终止启动的子进程. 直到子进程终止或者超时后强制杀死子进程后返回.
func (m *manager) terminateWorker(ctx context.Context) error {
	worker := m.Worker

	if worker.Cmd == nil || worker.Cmd.Process == nil {
		logger.Info("worker is not running")
		return nil
	}

	// 给子进程发送终止信号
	if err := sendTerminalSignal(worker.Cmd.Process); err != nil {
		return err
	}
	logger.Info("sent terminate signal to worker")

	waitCmd := make(chan error)
	go func() {
		waitCmd <- worker.Cmd.Wait()
	}()

	select {
	case err := <-waitCmd:
		logger.Info("worker terminated")
		return err
	case <-ctx.Done():
		// 超时后, 强制杀死子进程
		_ = worker.Cmd.Process.Kill()
		logger.Info("timeout, force kill worker")
		return nil
	}
}

// upgradeWorker 只是组合了 terminateWorker 和 runWorker 的逻辑. 先终止当前子进程, 再启动新的子进程.
func (m *manager) upgradeWorker(ctx context.Context) error {
	// 标记状态为重启中
	m.Status = managerStatusRestarting
	logger.Info("upgrading worker\n")

	terminateCxt, cancel := context.WithTimeout(ctx, m.Config.TerminateTimeout)
	defer cancel()
	// 终止当前子进程
	if err := m.terminateWorker(terminateCxt); err != nil {
		return err
	}

	// 启动新的子进程
	if err := m.runWorker(ctx); err != nil {
		return err
	}

	// 标记状态为运行中
	m.Status = managerStatusRunning
	logger.Info("worker upgraded\n")

	return nil
}

// checkNewVersion 会检查是否有新版本可用, 如果有, 则会调用 upgradeWorker 升级.
func (m *manager) checkNewVersion(ctx context.Context, fetcher fetcher.Fetcher) error {
	// 正在重启时, 跳过.
	if m.Status == managerStatusRestarting {
		return nil
	}

	logger.Info("checking new version")

	reader, err := fetcher.Fetch()
	if err != nil {
		return err
	}
	if reader == nil {
		return nil
	}

	// 可以的话, 关闭 reader
	if closer, ok := reader.(io.ReadCloser); ok {
		defer func() {
			_ = closer.Close()
		}()
	}

	tempFile, err := creteTempFile()
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	// 将新版本写入临时文件
	logger.Info("writing new version to temp file")
	if _, err := io.Copy(tempFile, reader); err != nil {
		return err
	}

	// 比较新版本的 hash 值和当前进程的 hash 值, 如果相同, 则跳过.
	hash, err := utils.MD5File(tempFile.Name())
	if err != nil {
		return err
	}
	if hash == m.BinHash {
		logger.Info("hash equal, no new version")
		return nil
	}

	// 继承权限
	if err := tempFile.Chmod(m.BinPermissions); err != nil {
		return err
	}

	// 继承 uid, gid
	if err := chown(tempFile, uid, gid); err != nil {
		return err
	}

	if err := m.Config.OnBeforeUpgrade(); err != nil {
		logger.Warn("upgrade canceled by OnBeforeUpgrade")
		return err
	}

	// 关闭临时文件, 避免冲突.
	logger.Info("closing temp file to avoid conflict")
	if err := tempFile.Close(); err != nil {
		return err
	}

	// 替换可执行文件.
	logger.Info("replacing executable file")
	if err := replaceExecutableFile(tempFile.Name(), m.BinPath); err != nil {
		return err
	}

	logger.Info("replaced executable file, upgrading")

	m.BinHash = hash

	// 重启子进程
	if err := m.upgradeWorker(ctx); err != nil {
		return fmt.Errorf("upgrade worker failed: %w", err)
	}

	return nil
}

// manager 将通过 workerProxy 来管理 worker 进程.
type workerProxy struct {
	Cmd *exec.Cmd
}