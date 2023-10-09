package fetcher

import "io"

// Func converts a fetch function into the fetcher interface
func Func(fn func(includeFile bool) (*AssetInfo, io.ReadCloser, FetchedBinaryUsedCallback, error)) Fetcher {
	return &fetcher{fn}
}

type fetcher struct {
	fn func(includeFile bool) (*AssetInfo, io.ReadCloser, FetchedBinaryUsedCallback, error)
}

func (f fetcher) Init() error {
	return nil //skip
}

func (f fetcher) Fetch(includeFile bool) (*AssetInfo, io.ReadCloser, FetchedBinaryUsedCallback, error) {
	return f.fn(includeFile)
}

func (f fetcher) GetName() string {
	return "Func Fetcher"
}
