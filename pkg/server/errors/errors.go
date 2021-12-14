package errors

import (
	"net/http"

	"git.backbone/corpix/unregistry/pkg/errors"
)

type Error struct {
	Code int
	Text string
	Err  error
	Meta interface{}
}

func (e *Error) Error() string {
	if e.Text == "" {
		return http.StatusText(e.Code)
	}

	return e.Text
}

func (e *Error) Chain() error {
	if e.Err != nil {
		return errors.Wrap(e.Err, e.Error())
	}

	return e
}

func NewError(code int, text string, err error, meta interface{}) *Error {
	return &Error{
		Code: code,
		Text: text,
		Err:  err,
		Meta: meta,
	}
}
