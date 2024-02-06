package overseer

import (
	"github.com/go-errors/errors"
	psuProc "github.com/shirou/gopsutil/v3/process"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/net/context"
	"net"
	"os"
)

type workerStatus int

const (
	// workerStatusInitialing worker 正在初始化
	workerStatusInitialing workerStatus = iota
	// workerStatusTaking worker 正在接管
	workerStatusTasking
	workerStatusDestroyed
)

type worker struct {
	Program ProgramFunc
	Config  Config
	Status  workerStatus
	// BinPath 用于标识当前进程的二进制文件的路径.
	BinPath string
	// BinHash 用于标识当前进程的二进制文件的 hash 值.
	BinHash   string
	listeners []net.Listener
}

// Initial 初始化 worker 参数并监听 Config 中指定的网络地址然后启动 Config.Program.
func (w worker) Initial(ctx context.Context) error {
	logger.Info("initializing worker")

	// 设置初始状态值
	w.Status = workerStatusInitialing

	path, err := os.Executable()
	if err != nil {
		return errors.New(err)
	}

	// 获取当前进程的二进制文件的路径
	w.BinPath = path

	// 获取当前进程的二进制文件的 hash 值
	if hash, err := utils.MD5File(path); err != nil {
		return errors.New(err)
	} else {
		w.BinHash = hash
	}

	envShortHash := os.Getenv(envShortBinHash)
	if len(envShortHash) <= 0 {
		return errors.Errorf("env short hash %s is empty", envShortBinHash)
	} else if envShortHash != w.BinHash[len(w.BinHash)-8:] {
		return errors.Errorf("hash mismatch: %s != %s", envShortHash, w.BinHash[len(w.BinHash)-8:])
	}

	managerProc, err := psuProc.NewProcess(int32(os.Getpid()))
	running, err := managerProc.IsRunning()
	if err != nil {
		return errors.New(err)
	} else if !running {
		return errors.Errorf("manager process is not running: %w", err)
	}

	// windows 不支持进程间信号通知.
	//// 通知 manager, worker 已经初始化完成
	//logger.Info("notifying manager that worker is ready")
	//if managerProcess, err := os.FindProcess(int(managerProc.Pid)); err != nil {
	//	return errors.Errorf("failed to find manager process: %w", err)
	//} else if err := managerProcess.Signal(InitialCompleteSignal); err != nil {
	//	return errors.Errorf("failed to signal manager: %w", err)
	//}
	//
	//// 等待 manager 发送接管信号
	//logger.Info("waiting for manager to signal take over")
	//signals := make(chan os.Signal)
	//signal.Notify(signals, TakeOverSignal)
	//<-signals

	// 监听 Config.Addresses 中指定的网络地址
	logger.Info("get take over signal, listening on addresses")
	if listeners, err := utils.ListenTCPAddresses(w.Config.Addresses...); err != nil {
		return errors.New(err)
	} else {
		w.listeners = listeners
	}

	// 设置状态为 workerStatusTasking
	w.Status = workerStatusTasking

	// 启动 Config.Program, 并保存返回的退出回调函数.
	logger.Info("starting program")
	if err := w.Program(ProgramProps{Listeners: w.listeners}); err != nil {
		return errors.New(err)
	}

	logger.Info("worker initialized")

	return nil
}

// Destroy 将关闭所有的打开的监听器并调用 programExit.
func (w worker) Destroy(ctx context.Context) {
	logger.Info("destroying worker")

	w.Status = workerStatusDestroyed

	// 关闭所有的监听器
	logger.Info("closing listeners")
	for _, listener := range w.listeners {
		_ = listener.Close()
	}

	logger.Info("worker destroyed")
}
