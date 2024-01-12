package runner

import (
	"bufio"
	"github.com/teivah/broadcast"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	git2 "github.com/siaikin/home-dashboard/internal/app/cron_service/git"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
)

var logger = comfy_log.New("[cron_service runner]")

type NodejsRunner struct {
	// Bin Nodejs 二进制文件路径
	Bin     string
	Project model.Project
	relay   *broadcast.Relay[string]
	C       chan string
}

func (r *NodejsRunner) Run(branch string) error {
	defer func() {
		r.relay.Close()
	}()
	npmBin := filepath.Join(r.Bin, "npm")

	runDir := constants.ProjectRunPath(r.Project, branch+lo.RandomString(16, lo.AlphanumericCharset))
	repoPath := constants.RepositoryPath(r.Project)

	if err := git2.CloneAndCheckout(repoPath, runDir, branch); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(runDir); err != nil {
			logger.Error("remove run dir error: %v", err)
		}
	}()

	outBuffer := strings.Builder{}
	errBuffer := strings.Builder{}

	logger.Info("project %s(%s) run npm install", r.Project.Name, branch)
	cmd := exec.Command(npmBin, "install")
	cmd.Dir = runDir
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer
	if err := cmd.Run(); err != nil {
		logger.Error(`run npm install error: %v, out: %s, err: %s`, err, outBuffer.String(), errBuffer.String())
		return err
	}

	logger.Info("project %s(%s) run npm build", r.Project.Name, branch)
	outBuffer.Reset()
	errBuffer.Reset()
	cmd = exec.Command(npmBin, "run", "build")
	cmd.Dir = runDir
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)

	if err := cmd.Start(); err != nil {
		return err
	}

	for scanner.Scan() {
		data := scanner.Text()
		//r.relay.Notify(data)
		r.C <- data
	}

	close(r.C)

	if err := cmd.Wait(); err != nil {
		logger.Error(`run npm build error: %v, out: %s, err: %s`, err, outBuffer.String(), errBuffer.String())
		return err
	}

	return nil
}

func (r *NodejsRunner) Listener() *broadcast.Listener[string] {
	return r.relay.Listener(1)
}

func NewNodejsRunner(project model.Project, bin string) *NodejsRunner {
	return &NodejsRunner{
		Bin:     bin,
		Project: project,
		relay:   broadcast.NewRelay[string](),
		C:       make(chan string, 1),
	}
}
