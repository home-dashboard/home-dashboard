package fetcher

import (
	"encoding/gob"
	"io"
)

type Fetcher interface {
	Init() error
	// Fetch 用于获取程序的二进制文件, 并返回二进制文件的信息和内容.
	// 可以通过 includeFile 参数来控制是否下载二进制文件.
	Fetch(includeFile bool) (*AssetInfo, io.ReadCloser, FetchedBinaryUsedCallback, error)
	// GetName 返回 fetcher 的名称.
	GetName() string
}

type AssetInfo struct {
	FetcherName  string `json:"fetcherName,omitempty"`
	Version      string `json:"version,omitempty"`
	ReleaseNotes string `json:"releaseNotes,omitempty"`
	URL          string `json:"url,omitempty"`
}

func init() {
	gob.Register(AssetInfo{})
}

// FetchedBinaryUsedCallback 二进制文件确认使用后, 调用此回调函数以更新 fetcher 内部的状态.
type FetchedBinaryUsedCallback func()
