package fetcher

import (
	"fmt"
	"github.com/google/go-github/v50/github"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"golang.org/x/mod/semver"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var logger = comfy_log.New("[overseer fetcher github]")

type client struct {
	*github.Client
}

var httpClient *client

// 初始化 Github 的 http client
func clientInitial(token string) error {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tokenClient := oauth2.NewClient(context.Background(), ts)

	// GitHub 的 http client
	httpClient = &client{github.NewClient(tokenClient)}

	return nil
}

var eTag string

// GetLatestReleaseByETag 获取最新的发布版本, 并且根据 ETag 判断是否需要更新.
// 如果 ETag 匹配, 则返回的 release 为 nil, 返回的 resp 状态码为 304 Not Modified.
// 否则, 返回的 release 中包含最新的发布版本.
func (c *client) GetLatestReleaseByETag(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	u := fmt.Sprintf("repos/%s/%s/releases/latest", owner, repo)

	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// 设置 If-None-Match 请求头.
	// 如果 ETag 匹配, 则返回 304 Not Modified 响应, 这时候就不需要更新了.
	// 否则, 返回 200 OK 响应, 并且返回最新的发布版本.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag#caching_of_unchanged_resources
	// https://docs.github.com/en/rest/activity/notifications?apiVersion=2022-11-28#about-github-notifications
	if len(eTag) > 0 {
		req.Header.Add("If-None-Match", eTag)
	}

	release := new(github.RepositoryRelease)
	if resp, err := c.Do(ctx, req, release); err != nil {
		// 如果返回 304 Not Modified 响应, 则不需要更新.
		if resp != nil && resp.StatusCode == http.StatusNotModified {
			return nil, resp, nil
		} else {
			return nil, nil, err
		}
	} else {
		// 获取 Last-Modified 响应头.
		eTag = resp.Header.Get("ETag")

		return release, resp, nil
	}
}

type GitHubFetcher struct {
	// GitHub 个人访问令牌
	Token string `json:"token"`
	// 当前版本
	CurrentVersion string `json:"currentVersion"`
	// GitHub 用户名和仓库名
	User, Repository string
	// SelectAsset 用于查找匹配的发布资产. 默认情况下, 如果它包含 GOOS 和 GOARCH, 则文件将匹配.
	SelectAsset func(filename string) bool `json:"selectAsset"`
	// SelectBinary 用于选择要使用的二进制文件. 默认情况下, 它将选择第一个找到的与当前操作系统和架构匹配的文件.
	SelectBinary func(filename string) bool `json:"selectBinary"`
	// OnProgress 用于跟踪下载进度.
	OnProgress func(written, total uint64) `json:"onProgress"`

	// latestETag
	latestETag        string
	latestReleaseInfo *github.RepositoryRelease
}

// Init validates the provided config
func (h *GitHubFetcher) Init() error {
	if httpClient == nil {
		if err := clientInitial(h.Token); err != nil {
			return err
		}
	}

	if h.User == "" {
		return fmt.Errorf("user required")
	}
	if h.Repository == "" {
		return fmt.Errorf("repo required")
	}
	if h.SelectAsset == nil {
		h.SelectAsset = defaultSelectAsset
	}
	if h.SelectBinary == nil {
		h.SelectBinary = defaultSelectBinary
	}
	if h.OnProgress == nil {
		h.OnProgress = defaultOnProgress
	}

	return nil
}

// Fetch 从提供的仓库中获取二进制文件. 如果 includeFile 为 false, 则只返回 AssetInfo.
func (h *GitHubFetcher) Fetch(includeFile bool) (*AssetInfo, io.ReadCloser, FetchedBinaryUsedCallback, error) {
	assetInfo := &AssetInfo{
		FetcherName: h.GetName(),
	}
	release, resp, err := httpClient.GetLatestReleaseByETag(context.Background(), h.User, h.Repository)
	if err != nil {
		return nil, nil, nil, err
	} else {
		if resp.StatusCode == http.StatusNotModified {
			release = h.latestReleaseInfo
		} else if release != nil {
			h.latestReleaseInfo = release
		}

	}
	assetInfo.Version = release.GetTagName()
	assetInfo.ReleaseNotes = release.GetBody()
	assetInfo.URL = release.GetHTMLURL()

	// 比较版本号, 如果
	// 1 当前版本号或 release 版本号非法
	// 2 当前版本号大于等于 release 版本号
	// 则不需要更新.
	if !semver.IsValid(h.CurrentVersion) {
		logger.Info("current version %s is invalid, not need to update", h.CurrentVersion)
		return nil, nil, nil, nil
	} else if !semver.IsValid(assetInfo.Version) {
		logger.Info("release version %s is invalid, not need to update", assetInfo.Version)
		return nil, nil, nil, nil
	} else if semver.Compare(h.CurrentVersion, assetInfo.Version) >= 0 {
		logger.Info("current version %s, release version %s. no need to update", h.CurrentVersion, assetInfo.Version)
		return nil, nil, nil, nil
	}

	if !includeFile {
		return assetInfo, nil, nil, nil
	}

	matchedAsset := h.findMatchingAsset(release)
	if matchedAsset == nil {
		return assetInfo, nil, nil, nil
	}

	// 创建临时目录
	tempDir, err := utils.CreateTempDir("fetch_from_github")
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	if err := h.downloadAndExtractArchive(matchedAsset, tempDir); err != nil {
		return nil, nil, nil, err
	}

	selectedFile, err := h.findSelectedFile(tempDir)
	if err != nil {
		return nil, nil, nil, err
	}

	// 读取二进制文件
	if file, err := utils.FileOpenOnlyFile(selectedFile); err != nil {
		return nil, nil, nil, err
	} else {
		return assetInfo, file, func() {
			h.CurrentVersion = assetInfo.Version
		}, nil
	}
}

func (h *GitHubFetcher) GetName() string {
	return "GitHub Fetcher"
}

// findMatchingAsset 查找匹配当前操作系统和架构的 [github.ReleaseAsset]
func (h *GitHubFetcher) findMatchingAsset(release *github.RepositoryRelease) *github.ReleaseAsset {
	for _, asset := range release.Assets {
		if h.SelectAsset(asset.GetName()) {
			return asset
		}
	}
	return nil
}

// findSelectedFile 从匹配的 [github.ReleaseAsset] 的解压文件中查找需要的二进制文件
func (h *GitHubFetcher) findSelectedFile(tempDir string) (string, error) {
	var selectedFile string
	walkFunc := func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if h.SelectBinary(d.Name()) {
			selectedFile = path
			return filepath.SkipDir
		}

		return nil
	}
	if err := filepath.WalkDir(tempDir, walkFunc); err != nil {
		return "", err
	}
	if selectedFile == "" {
		return "", fmt.Errorf("no binary found in %s", tempDir)
	}
	return selectedFile, nil
}

func (h *GitHubFetcher) downloadAndExtractArchive(matchedAsset *github.ReleaseAsset, extractPath string) error {
	// 创建临时文件
	tempFile, err := utils.CreateTempFile(matchedAsset.GetName())
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	_, assetUrl, err := httpClient.Repositories.DownloadReleaseAsset(context.Background(), h.User, h.Repository, matchedAsset.GetID(), nil)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", assetUrl, nil)
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
	if err := utils.CopyHttpResponseWithProgress(resp, tempFile, h.OnProgress); err != nil {
		return err
	}

	//resp, err := os.Open("D:\\projects\\go_projects\\home-dashboard\\dist\\home-dashboard_1.2.0_Windows_x86_64.zip")
	//if err != nil {
	//	return err
	//}
	//stat, err := resp.Stat()
	//if err != nil {
	//	return err
	//}
	//if err := utils.CopyWithProgress(resp, uint64(stat.Size()), tempFile, h.OnProgress); err != nil {
	//	return err
	//}

	// 解压临时文件到临时目录
	if err := utils.DecompressFile(tempFile.Name(), extractPath); err != nil {
		return err
	}

	return nil
}

func defaultSelectAsset(assetName string) bool {
	assetName = strings.ToLower(assetName)
	assetName = strings.Replace(assetName, "x86_64", "amd64", -1)
	return strings.Contains(assetName, runtime.GOOS) && strings.Contains(assetName, runtime.GOARCH)
}

func defaultSelectBinary(filename string) bool {
	filename = strings.ToLower(filename)

	switch filepath.Ext(filename) {
	case "":
	case ".exe":
		return true
	}

	return false
}

func defaultOnProgress(written, total uint64) {
	// log progress
}
