package overseer

import (
	"fmt"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/overseer/fetcher"
	"golang.org/x/net/context"
	"net"
	"os"
	"time"
)

const (
	envIsWorkerProcess  = "OVERSEER_IS_WORKER_PROCESS"
	envIsManagerProcess = "OVERSEER_IS_MANAGER_PROCESS"
	//envNumFDs         = "OVERSEER_NUM_FDS"
	envShortBinHash = "OVERSEER_SHORT_BIN_HASH"
)

var (
	logger           *comfy_log.Logger
	IsWorkerProcess  = false
	IsManagerProcess = false
)

func init() {
	// 根据环境变量判断当前进程是 manager 进程还是 worker 进程
	if _IsWorkerProcess, ok := os.LookupEnv(envIsWorkerProcess); ok {
		IsWorkerProcess = _IsWorkerProcess == "true"
	} else if _IsManagerProcess, ok := os.LookupEnv(envIsManagerProcess); ok {
		IsManagerProcess = _IsManagerProcess == "true"
	} else {
		IsManagerProcess = true
	}

	if IsWorkerProcess && !IsManagerProcess {
		logger = comfy_log.New("[overseer worker]")
	} else {
		logger = comfy_log.New("[overseer manager]")
	}
}

type Config struct {
	// Required 默认情况下 Program 在子进程中运行失败时, 会回退到在主进程中运行. Required 为 true 时, 会阻止其回退行为.
	Required bool `json:"required"`
	// Addresses socket 监听地址列表. 详情见 [net/http.Server.Addr]
	Addresses []string `json:"addresses"`
	// TerminateSignal 重启信令. 默认为 overseer.TerminateSignal
	TerminateSignal os.Signal `json:"terminateSignal"`
	// TerminateTimeout 等待子进程结束的超时时间. 默认为 5 秒.
	// 如果子进程在超时时间内未结束, 则会发送 SIGKILL 信令给子进程.
	TerminateTimeout time.Duration `json:"terminateTimeout"`
	// FetchInterval 通过 Fetches 轮询获取配置信息的调用间隔. 默认为 10 分钟.
	FetchInterval time.Duration `json:"fetchInterval"`
	// FetchTimeout 通过 Fetches 轮询获取配置信息的超时时间. 默认为 10 分钟.
	// 注意: 当二进制包较大或者网络较慢时, 将导致下载二进制包的时间较长. 请酌情调整该参数.
	FetchTimeout time.Duration `json:"fetchTimeout"`
	// OnBeforeUpgrade 在升级前执行的回调函数. 如果回调函数返回错误, 则会阻止升级.
	OnBeforeUpgrade func() error      `json:"onBeforeUpgrade"`
	Fetchers        []fetcher.Fetcher `json:"fetchers"`
}

// ProgramFunc 是程序主函数的类型, 接收 ProgramProps 参数并返回一个函数. 返回的函数会在程序更新或者退出时执行.
type ProgramFunc func(props ProgramProps) error

// ProgramProps 用于传递给 Config.Program 的参数
type ProgramProps struct {
	Listeners []net.Listener
}

// Run 将检查传入的 Config 是否合法, 并执行传入的 program 函数. 该函数会阻塞当前进程直到接收到终止信号.
// - 如果当前进程是 worker 进程, 则会执行 program 函数, 并阻塞当前进程, 直到接收到 config 中设置的 Config.TerminateSignal 信号.
// - 如果当前进程是 manager 进程, 则会执行 manager 函数, 并阻塞当前进程, 直到接收到系统中断信号.
func Run(program ProgramFunc, config Config) (func(ctx context.Context), error) {
	if validateConfig(&config) != nil {
		return nil, fmt.Errorf("invalid config: %w", validateConfig(&config))
	}
	setDefaultConfig(&config)

	if IsWorkerProcess && !IsManagerProcess {
		logger.Info("run worker with config: %+v", config)
		return runWorker(program, config)
	} else {
		logger.Info("run manager with config: %+v", config)
		return runManager(config)
	}
}

func validateConfig(config *Config) error {
	if config.Addresses == nil {
		return fmt.Errorf("addresses required")
	}
	return nil
}

// 对 Config 中的空参数设置默认值
func setDefaultConfig(config *Config) {
	if config.TerminateSignal == nil {
		config.TerminateSignal = TerminateSignal
	}
	if config.TerminateTimeout <= 0 {
		config.TerminateTimeout = time.Second * 5
	}
	if config.FetchInterval <= 0 {
		config.FetchInterval = time.Minute * 10
	}
	if config.FetchTimeout <= 0 {
		config.FetchTimeout = time.Minute * 10
	}
	if config.OnBeforeUpgrade == nil {
		config.OnBeforeUpgrade = func() error { return nil }
	}
	if config.Fetchers == nil {
		config.Fetchers = []fetcher.Fetcher{}
	}
}

func runWorker(program ProgramFunc, config Config) (func(ctx context.Context), error) {
	worker := worker{Program: program, Config: config}
	if err := worker.Initial(context.Background()); err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		worker.Destroy(ctx)
	}, nil
}

func runManager(config Config) (func(ctx context.Context), error) {
	manager := manager{Config: config}
	if err := manager.Initial(context.Background()); err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		manager.Destroy(ctx)

	}, nil
}
