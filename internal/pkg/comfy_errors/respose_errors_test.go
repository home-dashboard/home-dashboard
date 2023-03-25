package comfy_errors

import (
	"testing"
)

func TestNewResponseError(t *testing.T) {
	err := NewResponseError(UnknownError, "test")
	if err.Code != UnknownError {
		t.Errorf("expected code %d, got %d", UnknownError, err.Code)
	}

	if err.Err.Error() != "test" {
		t.Errorf("expected message %s, got %s", "test", err.Err.Error())
	}
}

func TestResponseError_Unwrap(t *testing.T) {
	err := NewResponseError(UnknownError, "test")

	if err.Unwrap().Error() != "test" {
		t.Errorf("expected message %s, got %s", "test", err.Err.Error())
	}

	if err.Unwrap().Error() != err.Err.Error() {
		t.Errorf("expected message %s, got %s", err.Err.Error(), err.Unwrap().Error())
	}
}
