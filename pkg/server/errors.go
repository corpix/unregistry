package server

import (
	"net/http"

	echo "github.com/labstack/echo/v4"

	serverErrors "git.backbone/corpix/unregistry/pkg/server/errors"
)

type (
	HTTPError = echo.HTTPError
	Error     = serverErrors.Error
)

var (
	ErrServerClosed = http.ErrServerClosed
)

func DefaultHTTPErrorHandler(err error, c Context) {
	if _, ok := err.(*HTTPError); ok {
		c.Echo().DefaultHTTPErrorHandler(err, c)
		return
	}

	//

	r := NewErrorResult(StatusInternalServerError, "")

	if e, ok := err.(*Error); ok {
		r.Error.Code = e.Code
		r.Error.Message = e.Error()
	}

	_ = c.JSON(r.Error.Code, r)
}

var NewError = serverErrors.NewError
