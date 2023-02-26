package comfy_errors

import "encoding/json"

type ResponseError struct {
	Code    ResponseErrorCode `json:"code"`
	Message string            `json:"message"`
}

// https://go.dev/doc/faq#methods_on_values_or_pointers
func (r ResponseError) Error() string {
	return r.Message
}

func (r ResponseError) String() string {
	marshal, err := json.Marshal(r)
	if err != nil {
		return ""
	}

	return string(marshal)
}

func NewResponseError(code ResponseErrorCode, message string) ResponseError {
	return ResponseError{
		Code:    code,
		Message: message,
	}
}
