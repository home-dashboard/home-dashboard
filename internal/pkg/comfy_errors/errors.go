package comfy_errors

import "encoding/json"

type ComfyError struct {
	Err comfyError `json:"err"`
}

func (e ComfyError) String() string {
	return e.Err.Error()
}

func (e ComfyError) Error() string {
	return e.String()
}

func (e ComfyError) JSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e ComfyError) Unwrap() error {
	return e.Err
}

type comfyError struct {
	error
}

// MarshalText 实现了 [encoding.TextMarshaler] 接口
func (e comfyError) MarshalText() ([]byte, error) {
	return []byte(e.Error()), nil
}
