package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/siaikin/home-dashboard/internal/app/cron_service/constants"
	"io"
	"path/filepath"
)

type Context struct {
	context.Context
	OverHTTP bool
}

func AdvertisedReferences(c Context, serviceType string, session transport.Session, w io.Writer) error {
	switch serviceType {
	case "git-upload-pack":
		return advertisedUploadPack(c, session.(transport.UploadPackSession), w)
	case "git-receive-pack":
		return advertisedReceivePack(c, session.(transport.ReceivePackSession), w)
	default:
		return fmt.Errorf("unsupported service %s", serviceType)
	}
}

func CreateSession(service, repositoryName string) (transport.Session, error) {
	switch service {
	case "git-upload-pack":
	case "git-receive-pack":
	default:
		return nil, fmt.Errorf("unsupported service %s", service)
	}

	repositoryPath := filepath.Join(constants.RepositoriesPath, repositoryName)

	endpoint, err := transport.NewEndpoint("/")
	if err != nil {
		return nil, err
	}

	fs := osfs.New(repositoryPath)
	svr := server.NewServer(server.NewFilesystemLoader(fs))

	switch service {
	case "git-upload-pack":
		return svr.NewUploadPackSession(endpoint, nil)
	case "git-receive-pack":
		return svr.NewReceivePackSession(endpoint, nil)
	default:
		return nil, fmt.Errorf("unsupported service %s", service)
	}
}

func advertisedUploadPack(c Context, session transport.UploadPackSession, w io.Writer) error {
	advRefs, err := session.AdvertisedReferencesContext(c)
	if err != nil {
		return err
	}
	if c.OverHTTP {
		advRefs.Prefix = [][]byte{
			[]byte("# service=git-upload-pack"),
			pktline.Flush,
		}
	}
	err = advRefs.Capabilities.Add("no-thin")
	if err != nil {
		return err
	}

	return advRefs.Encode(w)
}

func advertisedReceivePack(c Context, session transport.ReceivePackSession, w io.Writer) error {
	advRefs, err := session.AdvertisedReferencesContext(c)
	if err != nil {
		return err
	}
	if c.OverHTTP {
		advRefs.Prefix = [][]byte{
			[]byte("# service=git-receive-pack"),
			pktline.Flush,
		}
	}
	err = advRefs.Capabilities.Add("no-thin")
	if err != nil {
		return err
	}

	return advRefs.Encode(w)
}

func UploadPack(c Context, session transport.UploadPackSession, r io.Reader, w io.Writer) error {
	return uploadPack(c, session, r, w)
}
func uploadPack(c Context, session transport.UploadPackSession, r io.Reader, w io.Writer) error {
	req := packp.NewUploadPackRequest()
	if err := req.Decode(r); err != nil {
		return err
	}

	res, err := session.UploadPack(c, req)
	if err != nil {
		return err
	}

	return res.Encode(w)
}

func ReceivePack(c Context, session transport.ReceivePackSession, r io.Reader, w io.Writer) error {
	return receivePack(c, session, r, w)
}
func receivePack(c Context, session transport.ReceivePackSession, r io.Reader, w io.Writer) error {
	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(r); err != nil {
		return err
	}

	res, err := session.ReceivePack(c, req)
	if err != nil {
		return err
	}

	return res.Encode(w)
}
