package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/siaikin/home-dashboard/internal/pkg/notification"
	"github.com/siaikin/home-dashboard/internal/pkg/overseer"
	"github.com/siaikin/home-dashboard/internal/pkg/verison_info"
	"net/http"
	"time"
)

type UpgradeRequest struct {
	FetcherName string `form:"fetcherName"`
	Version     string `form:"version"`
}

// Upgrade
// @Summary Upgrade
// @Description Upgrade
// @Tags Upgrade
// @Accept json
// @Produce json
// @Param fetcherName body string true "fetcherName"
// @Param version body string true "version"
// @Success 202 {string} string "Accepted"
// @Router /upgrade [post]
func Upgrade(context *gin.Context) {
	body := UpgradeRequest{}

	if err := context.ShouldBindJSON(&body); err != nil {
		respondUnknownError(context, err.Error())
		return
	}

	logger.Info("receive upgrade request, fetcherName: %s, version: %s\n", body.FetcherName, body.Version)

	context.Status(http.StatusAccepted)

	var inst *overseer.Overseer
	var err error
	if inst, err = overseer.Get(); err != nil {
		abortWithError(context, http.StatusInternalServerError, errors.Errorf("overseer get failed. %w", err))
		return
	}

	if err := inst.Upgrade(body.FetcherName); err != nil {
		abortWithError(context, http.StatusInternalServerError, errors.Errorf("upgrade failed. %w", err))
		return
	}

	go func() {
		timer := time.NewTimer(time.Nanosecond)
		statusText := ""

		for {
			select {
			case <-timer.C:
				if status, err := inst.Status(); err != nil {
					logger.Error("get upgrade status failed. %v\n", err)
				} else if statusText != status.Text {
					// 通知前端更新进度
					notification.Send(overseer.StatusMessageType, map[string]interface{}{"type": status.Type, "text": status.Text, "extra": status.Extra})

					statusText = status.Text

					if status.Type == overseer.StatusTypeRunning || status.Type == overseer.StatusTypeDestroyed {
						return
					}
				}

				timer.Reset(time.Second)
			}
		}
	}()
}

// Version 获取版本信息
// @Summary Version
// @Description Version
// @Tags Version
// @Accept json
// @Produce json
// @Success 200 {object} VersionInfo
// @Router /version [get]
func Version(context *gin.Context) {
	context.JSON(http.StatusOK, VersionInfo{
		Version: verison_info.Version,
		Commit:  verison_info.Commit,
		Date:    verison_info.Date.UnixMilli(),
	})
}

type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    int64  `json:"date"`
}
