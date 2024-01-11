package runner

import (
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	git2 "github.com/siaikin/home-dashboard/internal/app/cron_service/git"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"os/exec"
	"path/filepath"
)

type NodejsRunner struct {
	// Bin Nodejs 二进制文件路径
	Bin     string
	Project model.Project
}

func (r *NodejsRunner) Run(branch string) error {
	//nodejsBin := filepath.Join(r.Bin, "node")
	npmBin := filepath.Join(r.Bin, "npm")

	runDir := constants.ProjectRunPath(r.Project, branch)
	repoPath := constants.RepositoryPath(r.Project)

	if err := git2.CloneAndCheckout(repoPath, runDir, branch); err != nil {
		return err
	}

	cmd := exec.Command(npmBin, "install")
	cmd.Dir = runDir

	return nil
}
