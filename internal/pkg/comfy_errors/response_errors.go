package comfy_errors

import (
	"fmt"
)

type ResponseError struct {
	ComfyError
	Code ResponseErrorCode `json:"code"`
}

func NewResponseError(code ResponseErrorCode, message string, a ...any) ResponseError {
	err := comfyError{fmt.Errorf(message, a...)}

	return ResponseError{
		ComfyError: ComfyError{Err: err},
		Code:       code,
	}
}
