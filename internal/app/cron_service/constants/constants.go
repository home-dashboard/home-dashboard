package constants

import (
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"os"
	"path/filepath"

	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
)

var logger = comfy_log.New("[cron_service constants]")

// RootPath 是被 cron_service 管理的文件的根目录.
// - RepositoriesPath: git 仓库的根目录.
// - NodejsPath: nodejs 的解压目录.
// - SSHPrivateKeyPath: ssh 私钥文件的路径.
var (
	RootPath          string
	ProjectsPath      string
	RepositoriesPath  string
	NodejsPath        string
	SSHPrivateKeyPath string
)

func init() {
	RootPath = filepath.Join(utils.WorkspaceDir(), "cron_service")
	ProjectsPath = filepath.Join(RootPath, "projects")
	RepositoriesPath = filepath.Join(RootPath, "repos")
	NodejsPath = filepath.Join(RootPath, "nodejs")

	if homeDir, err := os.UserHomeDir(); err != nil {
		logger.Fatal("os.UserHomeDir failed, %w", err)
	} else {
		SSHPrivateKeyPath = filepath.Join(homeDir, ".ssh", "id_rsa")
	}
}

func RepositoryPath(project model.Project) string {
	return filepath.Join(ProjectsPath, project.Name, project.Name+".git")
}

func DatabasePath(project model.Project, branchName string) string {
	return filepath.Join(ProjectsPath, project.Name, "database", branchName+".db")
}