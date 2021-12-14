package server

import (
	"net/http"
)

type (
	Result struct {
		Ok      bool          `json:"ok"`
		Error   *ResultError  `json:"error,omitempty"`
		Payload ResultPayload `json:"payload,omitempty"`
	}

	ResultError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	ResultPayload = interface{}
)

func NewErrorResult(code int, message string) Result {
	if message == "" {
		message = http.StatusText(code)
	}
	return Result{
		Ok: false,
		Error: &ResultError{
			Code:    code,
			Message: message,
		},
	}
}
