package registry

import (
	"fmt"

	"git.backbone/corpix/unregistry/pkg/server"
)

type ErrNotFound struct {
	Subject string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("no such %q", e.Subject)
}

func NewErrNotFound(subject string) error {
	inner := ErrNotFound{
		Subject: subject,
	}
	return server.NewError(
		server.StatusNotFound,
		"not found",
		inner,
		nil,
	)
}
