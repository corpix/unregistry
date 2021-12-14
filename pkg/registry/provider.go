package registry

import (
	"io"
	"strings"

	"git.backbone/corpix/unregistry/pkg/errors"
)

type (
	Stream io.ReadCloser

	Provider interface {
		GetManifest(name string, reference string) (Stream, error)
		GetBlob(name string, digest string) (Stream, error)
	}

	ProviderName string
)

const (
	ProviderLocalName ProviderName = "local"
)

var ProviderNames = []ProviderName{
	ProviderLocalName,
}

//

type ProviderConfig struct {
	Type string `yaml:"type"`

	Local *LocalProviderConfig `yaml:"local"`
}

func (c *ProviderConfig) Default() {
loop:
	for {
		switch {
		default:
			break loop
		}
	}
}

func (c *ProviderConfig) Validate() error {
	if c.Type == "" {
		return errors.New("type should not be empty")
	}

	providerName := strings.ToLower(c.Type)
	found := false
	for _, name := range ProviderNames {
		if string(name) == providerName {
			found = true
			break
		}
	}
	if !found {
		return errors.Errorf(
			"unsupported provider %q, choose one of: %v",
			c.Type, ProviderNames,
		)
	}

	return nil
}

//

func NewProvider(c ProviderConfig) (Provider, error) {
	switch strings.ToLower(c.Type) {
	case string(ProviderLocalName):
		return NewLocalProvider(*c.Local)
	default:
		return nil, errors.Errorf(
			"unsupported provider %q, choose one of: %v",
			c.Type, ProviderNames,
		)
	}
}
