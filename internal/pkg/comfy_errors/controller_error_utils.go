package comfy_errors

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type controllerUtils struct {
}

func (C *controllerUtils) RespondEntityAlreadyExistError(c *gin.Context, message string, a ...any) {
	_ = c.AbortWithError(http.StatusBadRequest, NewResponseError(EntityAlreadyExistsError, message, a...))
}

func (C *controllerUtils) RespondUnknownError(c *gin.Context, message string, a ...any) {
	_ = c.AbortWithError(http.StatusBadRequest, NewResponseError(UnknownError, message, a...))
}

func (C *controllerUtils) RespondEntityValidationError(c *gin.Context, message string, a ...any) {
	_ = c.AbortWithError(http.StatusBadRequest, NewResponseError(EntityValidationError, message, a...))
}

func (C *controllerUtils) RespondEntityNotFoundError(c *gin.Context, message string, a ...any) {
	_ = c.AbortWithError(http.StatusBadRequest, NewResponseError(EntityNotFoundError, message, a...))
}

var ControllerUtils = &controllerUtils{}
