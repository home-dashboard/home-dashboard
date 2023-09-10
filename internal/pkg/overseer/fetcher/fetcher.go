package fetcher

import "io"

type Fetcher interface {
	Init() error
	Fetch() (io.ReadCloser, error)
}
