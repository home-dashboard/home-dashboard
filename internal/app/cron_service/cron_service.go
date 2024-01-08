package cron_service

import (
	"context"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/git"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gliderlabs/ssh"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/controller"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
)

type CornServiceContext struct {
	context.Context
	Router *gin.RouterGroup
}

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

func ServeSSH(listener *net.Listener) error {
	server := ssh.Server{
		//ChannelHandlers: map[string]ssh.ChannelHandler{
		//	"session": func(srv *ssh.Server, conn *ssh2.ServerConn, newChan ssh2.NewChannel, ctx ssh.Context) {
		//		ch, req, err := newChan.Accept()
		//		if err != nil {
		//			return
		//		}
		//		handleSSHSession(constants.RepositoriesPath, ch, req)
		//	},
		//},
		Handler: func(session ssh.Session) {
			//gitProtolVersion, ok := lo.Find(session.Environ(), func(item string) bool {
			//	item ==
			//})

			commandArgs := session.Command()
			serviceType := commandArgs[0]
			repoName := commandArgs[1]

			log.Printf("dir: %s", filepath.Join(constants.RepositoriesPath, repoName))

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
	if homeDir, err := os.UserHomeDir(); err != nil {
		return err
	} else if err := server.SetOption(ssh.HostKeyFile(filepath.Join(homeDir, ".ssh", "id_rsa"))); err != nil {
		return err
	}

	return server.Serve(*listener)
}
