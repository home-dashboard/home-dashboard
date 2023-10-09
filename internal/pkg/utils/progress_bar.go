package utils

import (
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
)

// CopyHttpResponseWithProgress 从 http.Response 中读取数据，并将其写入 destination io.Writer 中，同时 onProgress 会被调用以获取进度.
func CopyHttpResponseWithProgress(resp *http.Response, destination io.Writer, onProgress func(written, total uint64)) error {
	return CopyWithProgress(resp.Body, uint64(resp.ContentLength), destination, onProgress)
}

// CopyWithProgress 从 source io.Reader 中读取数据，并将其写入 destination io.Writer 中，同时 onProgress 会被调用以获取进度.
func CopyWithProgress(source io.Reader, totalLength uint64, destination io.Writer, onProgress func(written, total uint64)) error {
	eg := new(errgroup.Group)

	// 创建一个 io.Pipe，用于获取进度
	readFromPipe, writeToPipe := io.Pipe()
	defer func() {
		_ = readFromPipe.Close()
	}()

	// 将源 io.Reader 包装在一个 io.TeeReader 中，以便在读取数据时获取进度
	tee := io.TeeReader(source, writeToPipe)

	total := totalLength
	written := uint64(0)
	eg.Go(func() error {
		buf := make([]byte, 32*1024)
		for i := 0; true; i++ {
			n, err := readFromPipe.Read(buf)
			if err != nil {
				if err == io.EOF {
					written += uint64(n)
					// 出现了 io.EOF 并且不是读取完毕的状态时, 触发 onProgress
					if written != total {
						onProgress(written, total)
					}
					return nil
				} else {
					return err
				}
			}

			written += uint64(n)
			// 循环十次或者读取完毕时, 触发 onProgress
			if i%10 == 0 || written == total {
				onProgress(written, total)
			}
		}
		return nil
	})

	eg.Go(func() error {
		_, err := io.Copy(destination, tee)
		// 读取完毕后, 关闭 io.Pipe 以触发 io.Pipe.Read 返回 io.EOF
		_ = writeToPipe.Close()
		return err
	})

	return eg.Wait()
}
