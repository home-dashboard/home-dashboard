package comfy_errors

import (
	"fmt"
	"github.com/go-errors/errors"
)

type ResponseError struct {
	ComfyError
	Code ResponseErrorCode `json:"code"`
}

func NewResponseError(code ResponseErrorCode, message string, a ...any) ResponseError {
	err := comfyError{errors.Wrap(fmt.Errorf(message, a...), 1)}

	return ResponseError{
		ComfyError: ComfyError{err},
		Code:       code,
	}
}
