package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
)

func respondEntityAlreadyExistError(c *gin.Context, message string, a ...any) {
	_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.EntityAlreadyExistsError, message, a...))
}

func respondUnknownError(c *gin.Context, message string, a ...any) {
	_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.UnknownError, message, a...))
}

func respondEntityValidationError(c *gin.Context, message string, a ...any) {
	_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.EntityValidationError, message, a...))
}

func respondEntityNotFoundError(c *gin.Context, message string, a ...any) {
	_ = c.AbortWithError(http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.EntityNotFoundError, message, a...))
}
