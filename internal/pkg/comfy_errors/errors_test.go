package comfy_errors

import (
	"errors"
	"testing"
)

func TestComfyError_String(t *testing.T) {
	e := ComfyError{
		Err: comfyError{errors.New("test")},
	}
	if got := e.String(); got != "test" {
		t.Errorf("ComfyError.String() = %v, want %v", got, "test")
	}
}

func TestComfyError_JSON(t *testing.T) {
	e := ComfyError{
		Err: comfyError{errors.New("test")},
	}
	if got, _ := e.JSON(); string(got) != `{"err":"test"}` {
		t.Errorf("ComfyError.JSON() = %v, want %v", string(got), `{"err":"test"}`)
	}
}

func TestComfyError_Unwrap(t *testing.T) {
	e := ComfyError{
		Err: comfyError{errors.New("test")},
	}
	if got := e.Unwrap(); got.Error() != "test" {
		t.Errorf("ComfyError.Unwrap() = %v, want %v", got.Error(), "test")
	}
	if got := e.Unwrap(); got != e.Err {
		t.Errorf("ComfyError.Unwrap() = %v, want %v", got, e.Err)
	}
}

func TestComfyError_Error(t *testing.T) {
	e := ComfyError{
		Err: comfyError{errors.New("test")},
	}
	if got := e.Error(); got != "test" {
		t.Errorf("ComfyError.Error() = %v, want %v", got, "test")
	}
	if got := e.Error(); got != e.String() {
		t.Errorf("ComfyError.Error() = %v, want %v", got, e.String())
	}
}
