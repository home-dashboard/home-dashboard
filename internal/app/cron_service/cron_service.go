package cron_service

import (
	"errors"
	"log"
	"net"

	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/git"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"

	"github.com/gin-gonic/gin"
	"github.com/gliderlabs/ssh"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/controller"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
)

var logger = comfy_log.New("[cron_Service]")

// MountHTTPRouter 把 cron_service 的 http 路由端点挂载到给定的 router 上.
func MountHTTPRouter(router *gin.RouterGroup) error {
	if err := loadRouter(router); err != nil {
		return err
	}

	if err := loadModel(); err != nil {
		return err
	}

	return nil
}

func loadRouter(router *gin.RouterGroup) error {
	// nodejs 相关接口
	router.GET("/nodejs/version/list_proxy", controller.ListNodejsVersion)
	router.POST("/nodejs/install/:version", controller.InstallNodejs)
	router.GET("/nodejs/installed", controller.InstalledNodejsInfo)

	// model.Project 相关接口
	router.POST("/project/create", controller.CreateProject)
	router.GET("/project/list", controller.ListProjects)
	router.PUT("/project/update", controller.UpdateProject)
	router.DELETE("/project/delete", controller.DeleteProject)

	// git 相关接口
	router.GET("/:name/info/refs", git.HTTPInfoRefs)
	router.POST("/:name/git-upload-pack", git.HTTPUploadPack)
	router.POST("/:name/git-receive-pack", git.HTTPReceivePack)

	return nil
}

func loadModel() error {
	return model.MigrateModel()
}

// ServeSSH 在 listener 上启动 ssh 服务.
// TODO 应该有一个分发器, 根据不同的请求, 调用不同的处理函数. 但目前就先这样吧
func ServeSSH(listener net.Listener) error {
	server := ssh.Server{
		Handler: func(session ssh.Session) {
			// 检查是否是 git 协议 v2
			_, ok := lo.Find(session.Environ(), func(item string) bool {
				return item == "GIT_PROTOCOL=version=2"
			})
			if !ok {
				logger.Warn("not supported git protocol\n")
				_ = session.Exit(1)
			}

			commandArgs := session.Command()
			serviceType := commandArgs[0]
			repoName := commandArgs[1]

			switch serviceType {
			case "git-upload-pack":
				if err := git.SSHUploadPack(session.Context(), repoName, session, session); err != nil {
					_ = session.Exit(1)
				}
				return
			case "git-receive-pack":
				if err := git.SSHReceivePack(session.Context(), repoName, session, session); err != nil {
					_ = session.Exit(1)
				}
				return
			default:
				log.Printf("unhandled cmd: %s", serviceType)
			}
		},
	}

	// 读取用户目录下的私钥
	if err := server.SetOption(ssh.HostKeyFile(constants.SSHPrivateKeyPath)); err != nil {
		return err
	}

	go func() {
		if err := server.Serve(listener); err != nil {
			if errors.Is(err, ssh.ErrServerClosed) {
				logger.Info("err: %v, don't worry, it's normal\n", err)
			} else {
				logger.Fatal("start ssh server failed, %w\n", err)
			}
		}

	}()

	return nil
}
