package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
)

func respondEntityAlreadyExistError(c *gin.Context, message string, a ...any) {
	abortWithError(c, http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.EntityAlreadyExistsError, message, a...))
}

func respondUnknownError(c *gin.Context, message string, a ...any) {
	abortWithError(c, http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, message, a...))
}

func respondEntityValidationError(c *gin.Context, message string, a ...any) {
	abortWithError(c, http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.EntityValidationError, message, a...))
}

func respondLoginError(c *gin.Context, message string, a ...any) {
	abortWithError(c, http.StatusBadRequest, comfy_errors.NewResponseError(comfy_errors.LoginRequestError, message, a...))
}

func respondEntityNotFoundError(c *gin.Context, message string, a ...any) {
	abortWithError(c, http.StatusNotFound, comfy_errors.NewResponseError(comfy_errors.EntityNotFoundError, message, a...))
}

func abortWithError(c *gin.Context, code int, err error) {
	c.Status(code)
	_ = c.Error(err)
}
