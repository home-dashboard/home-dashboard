package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	git2 "github.com/siaikin/home-dashboard/internal/app/cron_service/git"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/model"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
)

var logger = comfy_log.New("[cron_service runner]")

// Data NodejsRunner 获取到的数据
type Data struct {
	Table  string           `json:"table"`
	Rows   []map[string]any `json:"rows"`
	Origin string           `json:"origin"`
}

type NodejsRunner struct {
	// Bin Nodejs 二进制文件路径
	Bin     string
	Project model.Project
	C       chan Data
}

func (r *NodejsRunner) Run(branch string) error {
	defer close(r.C)

	npmBin := filepath.Join(r.Bin, "npm")

	runDir := constants.ProjectRunPath(r.Project, branch)
	defer func() {
		if err := os.RemoveAll(runDir); err != nil {
			logger.Error("remove run dir error: %v", err)
		}
	}()

	repoPath := constants.RepositoryPath(r.Project)
	if err := git2.CloneAndCheckout(repoPath, runDir, branch); err != nil {
		return fmt.Errorf("clone and checkout error: %w", err)
	}

	outputFilePath := constants.ProjectOutputPath(runDir)
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

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
	cmd.Env = append(cmd.Environ(), "OUTPUT_FILE="+outputFilePath)
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	if err := cmd.Run(); err != nil {
		logger.Error(`run npm build error: %v, out: %s, err: %s`, err, outBuffer.String(), errBuffer.String())
		return err
	}

	scanner := bufio.NewScanner(outputFile)
	for scanner.Scan() {
		str := scanner.Text()

		var d Data
		err := json.Unmarshal([]byte(str), &d)
		if err != nil {
			fmt.Printf("unmarshal json error: %v\n", err)
		}
		d.Origin = str

		r.C <- d
	}

	return scanner.Err()
}

func NewNodejsRunner(project model.Project, bin string) *NodejsRunner {
	return &NodejsRunner{
		Bin:     bin,
		Project: project,
		C:       make(chan Data, 1),
	}
}
