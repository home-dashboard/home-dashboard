package overseer

import (
	"fmt"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/overseer/fetcher"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/net/context"
	"net"
	"net/rpc"
	"os"
	"strconv"
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

var overseer *Overseer

func New(config Config) *Overseer {
	overseer = &Overseer{Config: config}
	return overseer
}

func Get() (*Overseer, error) {
	if overseer == nil {
		return nil, fmt.Errorf("overseer is not initialized, please call overseer.New() first")
	}
	return overseer, nil
}

type Config struct {
	// Required 默认情况下 Program 在子进程中运行失败时, 会回退到在主进程中运行. Required 为 true 时, 会阻止其回退行为.
	Required bool `json:"required"`
	// Port rpc 服务端口, rpc 服务用于内部进程间通讯.
	Port uint `json:"port"`
	// Addresses socket 监听地址列表. 详情见 [net/http.Server.Addr]
	Addresses []string `json:"addresses"`
	// TerminateTimeout 等待子进程结束的超时时间. 默认为 5 秒.
	// 如果子进程在超时时间内未结束, 则会发送 SIGKILL 信令给子进程.
	TerminateTimeout time.Duration `json:"terminateTimeout"`
	// FetchInterval 通过 Fetches 轮询获取配置信息的调用间隔. 默认为 10 分钟.
	FetchInterval time.Duration `json:"fetchInterval"`
	// FetchTimeout 通过 Fetches 轮询获取配置信息的超时时间. 默认为 10 分钟.
	// 注意: 当二进制包较大或者网络较慢时, 将导致下载二进制包的时间较长. 请酌情调整该参数.
	FetchTimeout time.Duration `json:"fetchTimeout"`
	// OnBeforeUpgrade 在升级前执行的回调函数. 如果回调函数返回错误, 则会阻止升级.
	OnBeforeUpgrade func() error `json:"onBeforeUpgrade"`
	// OnNewVersionFind 在发现新版本时执行的回调函数. 第一个传入的参数为新版本的信息.
	OnNewVersionFind func(info *fetcher.AssetInfo) `json:"onNewVersionFind"`
	// Fetches 查询新版本信息. 目前可用的 fetcher 见 [fetcher]
	Fetchers []fetcher.Fetcher `json:"fetchers"`
}

// ProgramFunc 是程序主函数的类型, 接收 ProgramProps 参数并返回一个函数. 返回的函数会在程序更新或者退出时执行.
type ProgramFunc func(props ProgramProps) error

// ProgramProps 用于传递给 Config.Program 的参数
type ProgramProps struct {
	Listeners []net.Listener
}

type Overseer struct {
	Config    Config
	manager   *manager
	worker    *worker
	rpcServer *upgradeService
	rpcClient *upgradeServiceClient
}

// Run 将检查传入的 Config 是否合法, 并执行传入的 program 函数. 该函数会阻塞当前进程直到接收到终止信号.
// - 如果当前进程是 worker 进程, 则会执行 program 函数, 并阻塞当前进程, 直到接收到中断信号.
// - 如果当前进程是 manager 进程, 则会执行 manager 函数, 并阻塞当前进程, 直到接收到系统中断信号.
func (o *Overseer) Run(program ProgramFunc) (func(ctx context.Context), error) {
	if err := validateConfig(&o.Config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	setDefaultConfig(&o.Config)

	logger.Info("run with config: %+v", o.Config)

	if IsWorkerProcess && !IsManagerProcess {
		return o.runWorker(program)
	} else {
		return o.runManager()
	}
}

// Upgrade 使用传入的 fetcher 拉取最新的二进制文件升级.
func (o *Overseer) Upgrade(fetcherName string) error {
	if IsWorkerProcess && !IsManagerProcess {
		return o.rpcClient.Upgrade(fetcherName)
	} else {
		go func() {
			if err := o.manager.Upgrade(fetcherName); err != nil {
				logger.Error("upgrade failed: %v", err)
			}
		}()
		return nil
	}
}

// Status 查询当前升级状态.
func (o *Overseer) Status() (Status, error) {
	if IsWorkerProcess && !IsManagerProcess {
		return o.rpcClient.Status()
	} else {
		return o.manager.Status(), nil
	}
}

// LatestVersionInfo 获取最新版本的信息. 可能返回 nil, 如果还未找到最新版本的信息.
func (o *Overseer) LatestVersionInfo() (*fetcher.AssetInfo, error) {
	if IsWorkerProcess && !IsManagerProcess {
		return o.rpcClient.LatestVersionInfo()
	} else {
		if o.manager.LatestAssetInfo == nil {
			return nil, fmt.Errorf("no latest version info")
		} else {
			return o.manager.LatestAssetInfo, nil
		}
	}
}

func (o *Overseer) runWorker(program ProgramFunc) (func(ctx context.Context), error) {
	if err := o.initialRpcClient(); err != nil {
		return nil, err
	}

	o.worker = &worker{Program: program, Config: o.Config}
	if err := o.worker.Initial(context.Background()); err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		o.worker.Destroy(ctx)
	}, nil
}

func (o *Overseer) runManager() (func(ctx context.Context), error) {
	stopRpcServer, err := o.initialRpcServer()
	if err != nil {
		return nil, err
	}

	o.manager = &manager{Config: o.Config}
	if err := o.manager.Initial(context.Background()); err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		stopRpcServer()
		o.manager.Destroy(ctx)
	}, nil
}

func (o *Overseer) initialRpcClient() error {
	if o.rpcClient != nil {
		return fmt.Errorf("rpc client already initial")
	}

	if client, err := rpc.Dial("tcp", ":"+strconv.FormatInt(int64(o.Config.Port), 10)); err != nil {
		return err
	} else {
		o.rpcClient = &upgradeServiceClient{Client: client, OverseerInstance: o}
	}

	return nil
}

// initialRpcServer 注册 rpc 服务并在 Config.Port 端口上监听请求. 返回一个函数, 调用以关闭端口监听.
func (o *Overseer) initialRpcServer() (func(), error) {
	if o.rpcServer != nil {
		return nil, fmt.Errorf("rpc server already initial")
	}

	server := rpc.NewServer()
	o.rpcServer = &upgradeService{OverseerInstance: o}
	if err := server.RegisterName(upgradeServiceName, o.rpcServer); err != nil {
		return nil, err
	}

	listener, err := utils.ListenTCPAddress(":" + strconv.FormatInt(int64(o.Config.Port), 10))
	if err != nil {
		return nil, err
	}
	go server.Accept(listener)

	return func() {
		_ = listener.Close()
	}, nil
}

func validateConfig(config *Config) error {
	if config.Addresses == nil {
		return fmt.Errorf("addresses required")
	}
	return nil
}

// 对 Config 中的空参数设置默认值
func setDefaultConfig(config *Config) {
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
	if config.Port <= 0 {
		config.Port = 8081
	}
	if config.OnNewVersionFind == nil {
		config.OnNewVersionFind = func(info *fetcher.AssetInfo) {
			// no-op
		}
	}
}
