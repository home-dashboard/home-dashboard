package constants

import (
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"path/filepath"
)

// RootPath 是被 cron_service 管理的文件的根目录.
// - RepositoriesPath: git 仓库的根目录.
// - NodejsPath: nodejs 的解压目录.
var RootPath = filepath.Join(utils.WorkspaceDir(), "cron_service")
var RepositoriesPath = filepath.Join(RootPath, "repos")
var NodejsPath = filepath.Join(RootPath, "nodejs")
