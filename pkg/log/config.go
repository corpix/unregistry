package log

import (
	"git.backbone/corpix/unregistry/pkg/bus"
	"git.backbone/corpix/unregistry/pkg/errors"
)

var (
	DefaultConfig = Config{
		Level: "info",
	}
)

type Config struct {
	Level string
}

func (c *Config) Default() {
loop:
	for {
		switch {
		case c.Level == "":
			c.Level = DefaultConfig.Level
		default:
			break loop
		}
	}
}

func (c *Config) Validate() error {
	if c.Level == "" {
		return errors.New("level should not be empty")
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
