package nodejs

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/mod/semver"
)

type Platform string
type CPU string

const (
	WINDOWS Platform = "windows"
	LINUX   Platform = "linux"
	OSX     Platform = "darwin"

	i386  CPU = "i386"
	amd64 CPU = "amd64"
	arm64 CPU = "arm64"
	ppc64 CPU = "ppc64"

	ExtensionTarGz = "tar.gz"
	ExtensionZip   = "zip"
)

type Installer struct {
	MirrorURL string
	// WorkDirectory 是相对与当前工作目录的路径
	WorkDirectory string
	// OnProgress 用于跟踪下载进度.
	OnProgress func(written, total uint64)
}

func (r *Installer) ResolvePath(version string, platform string, cpu string) (string, error) {
	os := WINDOWS
	arch := ""
	extension := ExtensionZip

	switch platform {
	case string(WINDOWS):
		os = "win"
		extension = ExtensionZip
	case string(LINUX):
		os = LINUX
		extension = ExtensionTarGz
	case string(OSX):
		os = OSX
		extension = ExtensionTarGz
	default:
		return "", fmt.Errorf("unsupported OS %s", platform)
	}

	switch cpu {
	case string(i386):
		arch = "x86"
	case string(amd64):
		arch = "x64"
	case string(arm64):
		arch = string(arm64)
	default:
		return "", fmt.Errorf("unsupported Arch %s", cpu)
	}

	return fmt.Sprintf("node-%s-%s-%s.%s", version, os, arch, extension), nil
}

func (r *Installer) Install(version string, platform string, cpu string) error {
	fileName, err := r.ResolvePath(version, platform, cpu)
	if err != nil {
		return err
	}

	// 创建临时文件
	tempFile, err := utils.CreateTempFile(fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	req, err := http.NewRequest("GET", strings.Join([]string{r.MirrorURL, "dist", "v" + version, fileName}, "/"), nil)
	if err != nil {
		return err
	}
	req = req.WithContext(context.Background())
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 将下载的文件写入临时文件
	if err := utils.CopyHttpResponseWithProgress(resp, tempFile, r.OnProgress); err != nil {
		return err
	}

	// 解压临时文件到 WorkDirectory
	if err := utils.DecompressFileToDir(tempFile.Name(), filepath.Join(utils.WorkspaceDir(), r.WorkDirectory)); err != nil {
		return err
	}

	return nil
}

func (r *Installer) Uninstall() error {
	return os.RemoveAll(filepath.Join(utils.WorkspaceDir(), r.WorkDirectory))
}

func (r *Installer) IsInstalled() bool {
	return semver.IsValid(r.Version())
}

type Info struct {
	Version    string `json:"version,omitempty"`
	Arch       string `json:"arch,omitempty"`
	OS         string `json:"os,omitempty"`
	Executable string `json:"executable,omitempty"`
}

// Info 返回当前系统的 Node.js 信息.
func (r *Installer) Info() Info {
	info := Info{
		Version:    r.Version(),
		Arch:       runtime.GOARCH,
		OS:         runtime.GOOS,
		Executable: r.Executable(),
	}

	return info
}

func (r *Installer) Executable() string {
	dir := filepath.Join("D:\\projects\\go_projects\\home-dashboard\\bin", r.WorkDirectory)
	switch runtime.GOOS {
	case string(WINDOWS):
		return filepath.Join(dir, "node.exe")
	case string(LINUX):
		return filepath.Join(dir, "bin", "node")
	case string(OSX):
		return filepath.Join(dir, "bin", "node")
	default:
		return ""
	}
}

func (r *Installer) Version() string {
	cmd := exec.Command(r.Executable(), "-v")

	if bytes, err := cmd.Output(); err != nil {
		return err.Error()
	} else {
		return strings.TrimSpace(string(bytes))
	}
}
