package cache_control

import (
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"io/fs"
	"net/http"
	"strings"
)

var eTagLogger = comfy_log.New("[cache_control etag]")

// ETagMiddleware ETag 中间件
// 参考: https://github.com/gin-gonic/gin/issues/1222#issuecomment-486922636
func ETagMiddleware(fsys fs.FS) gin.HandlerFunc {
	etagStore := make(map[string]string, 100)

	eTagLogger.Info("init ETag store...")
	fileCount := int64(0)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if hash, err := utils.MD5FSFile(fsys, path); err != nil {
			return err
		} else {
			etagStore[path] = hash
		}
		fileCount++

		return nil
	})
	if err != nil {
		eTagLogger.Fatal("init ETag store failed: %w\n", errors.New(err))
	} else {
		eTagLogger.Info("init ETag store complete, %d files\n", fileCount)
	}

	return func(context *gin.Context) {
		path := strings.Trim(context.Request.URL.EscapedPath(), "/")
		etag, ok := etagStore[path]
		if !ok {
			context.Next()
			return
		}

		context.Header("ETag", etag)

		if requestETag := context.GetHeader("If-None-Match"); requestETag != "" {
			if etag == requestETag {
				context.AbortWithStatus(http.StatusNotModified)
				return
			}
		}

		context.Next()
	}
}
