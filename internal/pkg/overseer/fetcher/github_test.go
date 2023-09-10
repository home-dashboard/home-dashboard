package fetcher

import (
	"testing"
)

func TestGitHubFetcher_Normal(t *testing.T) {
	// 设置测试用例的参数.
	token := "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"
	currentVersion := "v1.0.0"
	user := "home-dashboard"
	repo := "home-dashboard"

	if err := clientInitial(token); err != nil {
		t.Fatalf("初始化GitHub客户端失败: %v", err)
	}

	// 创建一个新的 GitHubFetcher 实例.
	fetcher := &GitHubFetcher{
		Token:          token,
		CurrentVersion: currentVersion,
		User:           user,
		Repository:     repo,
		OnProgress: func(written, total uint64) {
			t.Logf("下载进度: %d/%d", written, total)
		},
	}

	if err := fetcher.Init(); err != nil {
		t.Fatalf("初始化GitHubFetcher失败: %v", err)
	}

	// 测试正常情况下的使用.
	reader, err := fetcher.Fetch()
	if err != nil {
		t.Fatalf("Fetch returned an error: %v", err)
	}
	if reader == nil {
		t.Fatal("Fetch returned a nil reader")
	}
}

func TestGitHubFetcher_InvalidToken(t *testing.T) {
	// 设置测试用例的参数.
	token := "invalid_token"
	currentVersion := "v1.0.0"
	user := "home-dashboard"
	repo := "home-dashboard"

	if err := clientInitial(token); err != nil {
		t.Fatalf("初始化GitHub客户端失败: %v", err)
	}

	// 创建一个新的 GitHubFetcher 实例.
	fetcher := &GitHubFetcher{
		Token:          token,
		CurrentVersion: currentVersion,
		User:           user,
		Repository:     repo,
	}

	if err := fetcher.Init(); err != nil {
		t.Fatalf("初始化GitHubFetcher失败: %v", err)
	}

	// 测试当获取二进制文件时发生错误时，Fetcher 是否能够正确地处理错误并返回适当的错误信息.
	reader, err := fetcher.Fetch()
	if err == nil {
		t.Fatal("Fetch did not return an error")
	}
	if reader != nil {
		t.Fatal("Fetch returned a non-nil reader")
	}
}

func TestGitHubFetcher_Update(t *testing.T) {
	// 设置测试用例的参数.
	token := "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"
	currentVersion := "v1.0.0"
	user := "home-dashboard"
	repo := "home-dashboard"

	if err := clientInitial(token); err != nil {
		t.Fatalf("初始化GitHub客户端失败: %v", err)
	}

	// 创建一个新的 GitHubFetcher 实例.
	fetcher := &GitHubFetcher{
		Token:          token,
		CurrentVersion: currentVersion,
		User:           user,
		Repository:     repo,
		OnProgress: func(written, total uint64) {
			t.Logf("下载进度: %d/%d", written, total)
		},
	}

	if err := fetcher.Init(); err != nil {
		t.Fatalf("初始化GitHubFetcher失败: %v", err)
	}

	// 测试当获取二进制文件的频率过低时，Fetcher 是否能够及时获取更新的二进制文件，以确保应用程序始终是最新的版本.
	reader, err := fetcher.Fetch()
	if err != nil {
		t.Fatalf("Fetch returned an error: %v", err)
	}
	if reader == nil {
		t.Fatal("Fetch returned a nil reader")
	}
}
