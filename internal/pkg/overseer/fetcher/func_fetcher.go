package fetcher

import "io"

// Func converts a fetch function into the fetcher interface
func Func(fn func() (io.ReadCloser, error)) Fetcher {
	return &fetcher{fn}
}

type fetcher struct {
	fn func() (io.ReadCloser, error)
}

func (f fetcher) Init() error {
	return nil //skip
}

func (f fetcher) Fetch() (io.ReadCloser, error) {
	return f.fn()
}
