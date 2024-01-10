package git

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"io"
)

func SSHUploadPack(c context.Context, repoName string, r io.Reader, w io.Writer) error {
	session, err := CreateSession("git-upload-pack", repoName)
	if err != nil {
		return err
	}

	C := Context{
		Context:  c,
		OverHTTP: false,
	}

	err = AdvertisedReferences(C, "git-upload-pack", session, w)
	if err != nil {
		return err
	}

	return UploadPack(C, session.(transport.UploadPackSession), r, w)
}

func SSHReceivePack(c context.Context, repoName string, r io.Reader, w io.Writer) error {
	session, err := CreateSession("git-receive-pack", repoName)
	if err != nil {
		return err
	}

	C := Context{
		Context:  c,
		OverHTTP: false,
	}

	err = AdvertisedReferences(C, "git-receive-pack", session, w)
	if err != nil {
		return err
	}

	return ReceivePack(C, session.(transport.ReceivePackSession), ReaderFakeCloser{r: r}, w)
}

type ReaderFakeCloser struct {
	r io.Reader
}

func (rfc ReaderFakeCloser) Read(data []byte) (int, error) {
	return rfc.r.Read(data)
}

func (rfc ReaderFakeCloser) Close() error {
	return nil
}