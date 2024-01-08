package cron_service

import (
	"context"
	"net"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gliderlabs/ssh"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/controller"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	ssh2 "golang.org/x/crypto/ssh"
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
	router.GET("/:name/info/refs", controller.InfoRefs)
	router.POST("/:name/git-upload-pack", controller.UploadPack)
	router.POST("/:name/git-receive-pack", controller.ReceivePack)

	return nil
}

func loadModel() error {
	return model.MigrateModel()
}

func ServeSSH(listener *net.Listener) error {
	server := ssh.Server{
		//PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
		//	ssh2.MarshalAuthorizedKey()
		//	ssh.ParseAuthorizedKey()
		//	ssh2.MarshalAuthorizedKey()
		//},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"session": func(srv *ssh.Server, conn *ssh2.ServerConn, newChan ssh2.NewChannel, ctx ssh.Context) {
				// TODO 设置静态的仓库目录路径 repoDir
				// TODO 根据 OpenSSH 规范 从 authorized_keys 中读取用户的公钥  authorized_keys 文件相对于 repoDir 保存
				// TODO 增加 public key 的管理接口
				// TODO bare repo 的创建时 初始化一些文件 如 README.md .gitignore package.json 等 还有 database.json 表示数据库的配置
				// TODO database.json 中的配置信息
				// TODO 使用 json schema 定义 database.json 的格式。json schema 从 go struct 中生成 使用
				repoDir := filepath.Join(utils.WorkspaceDir(), "cron_service", "repos")
				ch, req, err := newChan.Accept()
				if err != nil {
					return
				}
				handleSSHSession(repoDir, ch, req)
			},
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
