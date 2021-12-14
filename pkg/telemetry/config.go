package telemetry

import (
	"git.backbone/corpix/unregistry/pkg/bus"
	"git.backbone/corpix/unregistry/pkg/errors"
	"git.backbone/corpix/unregistry/pkg/server"
)

type Config struct {
	Enable bool           `yaml:"enable"`
	Addr   string         `yaml:"addr"`
	Path   string         `yaml:"path"`
	HTTP   *server.Config `yaml:"http"`
}

func (c *Config) Default() {
loop:
	for {
		switch {
		case c.Addr == "":
			c.Addr = "127.0.0.1:4280"
		case c.Path == "":
			c.Path = "/"
		case c.HTTP == nil:
			c.HTTP = &server.Config{}
		default:
			break loop
		}
	}
}

func (c *Config) Validate() error {
	if !c.Enable {
		return nil
	}
	if c.Path == "" {
		return errors.New("path should not be empty")
	}

	return nil
}

func (c *Config) Update(cc interface{}) error {
	bus.Config <- bus.ConfigUpdate{
		Subsystem: Subsystem,
		Config:    cc,
	}
	return nil
}
