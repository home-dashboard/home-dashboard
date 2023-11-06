package main

import (
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/configuration"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
	"github.com/siaikin/home-dashboard/internal/pkg/overseer"
	"github.com/siaikin/home-dashboard/internal/pkg/overseer/fetcher"
	"github.com/siaikin/home-dashboard/internal/pkg/verison_info"
	"golang.org/x/net/context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var logger = comfy_log.New("[server_monitor]")

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	config := configuration.Get()
	if config.ServerMonitor.Port == 0 {
		logger.Fatal("port %d is invalid\n", config.ServerMonitor.Port)
	}
	if config.ServerMonitor.Development.Enable == false {
		gin.SetMode(gin.ReleaseMode)
	}

	// 保存版本信息
	if date, err := time.Parse(time.RFC3339, date); err != nil {
		now := time.Now()
		logger.Warn("parse date [%s] failed. use current time(%s) instead. %v\n", date, now.Format(time.RFC3339), err)
		verison_info.Set(version, commit, now)
	} else {
		verison_info.Set(version, commit, date)
	}
}

func main() {
	logger.Info("version %s, commit %s, built at %s\n", version, commit, date)

	// 设置数据库文件路径
	database.SetSourceFilePath("home-dashboard.db")
	db := database.GetDB()

	// 生成 overseer 配置
	overseerConfig, err := makeOverseerConfig()
	if err != nil {
		logger.Fatal("makeOverseerConfig failed. %v\n", err)
	}

	overseerInstance := overseer.New(overseerConfig)

	// 通过 overseer 启动主程序
	termFunc, err := overseerInstance.Run(func(props overseer.ProgramProps) error {
		// 仅 worker 进程会启动程序
		if !overseer.IsWorkerProcess {
			logger.Fatal("current process is not worker process\n")
		}

		server_monitor.Initial(db)

		listeners := props.Listeners
		if len(listeners) <= 0 {
			return fmt.Errorf("no listeners")
		}

		server_monitor.Start(context.Background(), &listeners[0])

		return nil
	})
	if err != nil {
		logger.Fatal("overseer.Run failed. %v\n", err)
	}

	// 阻塞当前进程, 直到接收到中断信号
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)
	sig := <-signals
	logger.Info("receive signal: %s, terminate %s\n", sig, lo.Ternary(overseer.IsManagerProcess, "manager", "worker"))

	// 优雅地关闭进程, 5秒后将强制结束进程
	// ctx 用于通知服务器有5秒来结束当前正在处理的请求
	// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/notify-without-context/server.go
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	termFunc(ctx)
	if overseer.IsWorkerProcess {
		server_monitor.Stop(ctx)
	}
	<-ctx.Done()

	logger.Info("server stopped\n")
}

func makeOverseerConfig() (overseer.Config, error) {
	config := configuration.Get()
	overseerConfig := overseer.Config{
		Addresses: []string{":" + strconv.FormatInt(int64(config.ServerMonitor.Port), 10)},
	}

	updateConfig := config.ServerMonitor.Update
	if !updateConfig.Enable {
		return overseerConfig, nil
	}

	// 从配置文件中获取更新源
	var fetchers []fetcher.Fetcher

	// 配置 GitHubFetcher
	githubFetcherConfig := updateConfig.Fetchers.GitHub
	if lo.IsNotEmpty(githubFetcherConfig.Owner) && lo.IsNotEmpty(githubFetcherConfig.Repository) {
		gitHubPST, ok := lo.Coalesce(config.ServerMonitor.ThirdParty.GitHub.PersonalAccessToken, githubFetcherConfig.PersonalAccessToken)
		if !ok {
			logger.Warn("not any github personal access token found\n")
		}

		fetchers = append(fetchers, &fetcher.GitHubFetcher{
			Token:          gitHubPST,
			CurrentVersion: version,
			User:           githubFetcherConfig.Owner,
			Repository:     githubFetcherConfig.Repository,
			OnProgress: func(written, total uint64) {
				logger.Info("overseer: downloading %f%% (%s/%s)\n", float64(written)/float64(total)*100, humanize.Bytes(written), humanize.Bytes(total))
			},
		})

		logger.Info("overseer: github fetcher enabled\n")
	}

	overseerConfig.Fetchers = fetchers

	if updateConfig.FetchInterval != 0 {
		overseerConfig.FetchInterval = updateConfig.FetchInterval * time.Second
	}
	if updateConfig.FetchTimeout != 0 {
		overseerConfig.FetchTimeout = updateConfig.FetchTimeout * time.Second
	}
	if updateConfig.Port != 0 {
		overseerConfig.Port = updateConfig.Port
	}

	return overseerConfig, nil
}
