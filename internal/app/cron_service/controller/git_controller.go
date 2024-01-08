package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"log"
	"net/http"
	"path/filepath"
)

// InfoRefs git smart http 协议接口 (git-upload-pack).
// @Summary git smart http 协议接口 (git-upload-pack).
// @Description git smart http 协议接口 (git-upload-pack).
// @Tags Git
// @Accept  json
// @Produce  json
// @Param name path string true "Repository name"
// @Success 200 {string} string
// @Router /{name}/info/refs [get]
func InfoRefs(c *gin.Context) {
	name := filepath.Clean(c.Param("name"))

	dir := filepath.Join(utils.WorkspaceDir(), "cron_service", "repos", name)
	r := c.Request
	rw := c.Writer

	log.Printf("httpInfoRefs %s %s", r.Method, r.URL)

	service := r.URL.Query().Get("service")
	if service != "git-upload-pack" && service != "git-receive-pack" {
		http.Error(rw, "only smart git", 403)
		return
	}

	rw.Header().Set("content-type", fmt.Sprintf("application/x-%s-advertisement", service))

	ep, err := transport.NewEndpoint("/")
	if err != nil {
		http.Error(rw, err.Error(), 500)
		log.Println(err)
		return
	}
	bfs := osfs.New(dir)
	ld := server.NewFilesystemLoader(bfs)
	svr := server.NewServer(ld)

	var sess transport.Session

	if service == "git-upload-pack" {
		sess, err = svr.NewUploadPackSession(ep, nil)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
	} else {
		sess, err = svr.NewReceivePackSession(ep, nil)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
	}

	ar, err := sess.AdvertisedReferencesContext(r.Context())
	if err != nil {
		http.Error(rw, err.Error(), 500)
		log.Println(err)
		return
	}
	ar.Prefix = [][]byte{
		[]byte(fmt.Sprintf("# service=%s", service)),
		pktline.Flush,
	}

	if err := ar.Capabilities.Add("no-thin"); err != nil {
		http.Error(rw, err.Error(), 500)
		log.Println(err)
		return
	}

	err = ar.Encode(rw)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		log.Println(err)
		return
	}
}

func UploadPack(c *gin.Context) {
	name := filepath.Clean(c.Param("name"))

	repo := filepath.Join(utils.WorkspaceDir(), "cron_service", "repos", name)

	c.Header("content-type", "application/x-git-upload-pack-result")

	upr := packp.NewUploadPackRequest()
	err := upr.Decode(c.Request.Body)
	if err != nil {
		comfy_errors.ControllerUtils.RespondEntityValidationError(c, err.Error())
		return
	}

	ep, err := transport.NewEndpoint("/")
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	billyfs := osfs.New(repo)
	loader := server.NewFilesystemLoader(billyfs)
	svr := server.NewServer(loader)
	session, err := svr.NewUploadPackSession(ep, nil)
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	res, err := session.UploadPack(c, upr)
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	if err = res.Encode(c.Writer); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}
}

func ReceivePack(c *gin.Context) {
	name := filepath.Clean(c.Param("name"))

	repo := filepath.Join(utils.WorkspaceDir(), "cron_service", "repos", name)

	c.Header("content-type", "application/x-git-receive-pack-result")

	upr := packp.NewReferenceUpdateRequest()
	err := upr.Decode(c.Request.Body)
	if err != nil {
		comfy_errors.ControllerUtils.RespondEntityValidationError(c, err.Error())
		return
	}

	ep, err := transport.NewEndpoint("/")
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	billyfs := osfs.New(repo)
	loader := server.NewFilesystemLoader(billyfs)
	svr := server.NewServer(loader)
	session, err := svr.NewReceivePackSession(ep, nil)
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	res, err := session.ReceivePack(c, upr)
	if err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}

	if err = res.Encode(c.Writer); err != nil {
		comfy_errors.ControllerUtils.RespondUnknownError(c, err.Error())
		return
	}
}
