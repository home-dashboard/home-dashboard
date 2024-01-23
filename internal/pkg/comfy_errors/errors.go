package comfy_errors

import (
	"github.com/go-errors/errors"
)

type ComfyError struct {
	comfyError
}

func (e ComfyError) String() string {
	return e.Err.Error()
}

func (e ComfyError) Error() string {
	return e.Err.Error()
}

type comfyError struct {
	*errors.Error
}

// MarshalText 实现了 [encoding.TextMarshaler] 接口
func (e comfyError) MarshalText() ([]byte, error) {
	return []byte(e.ErrorStack()), nil
}
