package config

import (
	"net/url"
	"strings"
	"time"

	"github.com/corpix/revip"

	"git.backbone/corpix/unregistry/pkg/bus"
	"git.backbone/corpix/unregistry/pkg/log"
	"git.backbone/corpix/unregistry/pkg/meta"
	"git.backbone/corpix/unregistry/pkg/registry"
	"git.backbone/corpix/unregistry/pkg/telemetry"
)

const (
	Subsystem = "config"
)

var (
	EnvironPrefix = meta.EnvNamespace

	LocalPostprocessors = []revip.Option{
		revip.WithDefaults(),
		revip.WithValidation(),
	}
	InitPostprocessors = []revip.Option{
		revip.WithDefaults(),
	}
)

var (
	Unmarshaler = revip.YamlUnmarshaler
	Marshaler   = revip.YamlMarshaler
)

type Config struct {
	Log               *log.Config
	Telemetry         *telemetry.Config
	Registry          *registry.Config
	ShutdownGraceTime time.Duration
}

func (c *Config) Default() {
loop:
	for {
		switch {
		case c.Log == nil:
			c.Log = &log.Config{}
		case c.Telemetry == nil:
			c.Telemetry = &telemetry.Config{}
		case c.Registry == nil:
			c.Registry = &registry.Config{}
		case c.ShutdownGraceTime == 0:
			c.ShutdownGraceTime = 120 * time.Second
		default:
			break loop
		}
	}
}

func (c *Config) Update(cc interface{}) error {
	bus.Config <- bus.ConfigUpdate{
		Subsystem: Subsystem,
		Config:    cc,
	}
	return nil
}

//

func Postprocess(c interface{}) error {
	return revip.Postprocess(c, LocalPostprocessors...)
}

func Default() (*Config, error) {
	c := &Config{}
	err := revip.Postprocess(
		c,
		revip.WithDefaults(),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func Load(paths []string, postprocessors ...revip.Option) (*Config, error) {
	var (
		c   = &Config{}
		err error
	)

	l, err := log.Create(log.Config{Level: "info"})
	if err != nil {
		return nil, err
	}

	//

	if len(postprocessors) == 0 {
		errorHandler := revip.UpdatesFromEtcdErrorHandler(func(err error) {
			l.Error().
				Err(err).
				Msg("got an error while handling update from etcd")
		})

		postprocessors = append(postprocessors, LocalPostprocessors...)
		for _, path := range paths {
			if strings.HasPrefix(path, revip.SchemeEtcd+":") {
				u, err := url.Parse(path)
				if err != nil {
					return nil, err
				}

				e, err := revip.NewEtcdClient(path)
				if err != nil {
					return nil, err
				}
				postprocessors = append(
					postprocessors,
					revip.WithUpdatesFromEtcd(
						e,
						strings.TrimPrefix(u.Path, "/"),
						Unmarshaler,
						errorHandler,
					),
				)
			}
		}
	}

	loaders := make([]revip.Option, len(paths))
	for n, path := range paths {
		loaders[n], err = revip.FromURL(
			strings.TrimSpace(path),
			Unmarshaler,
		)
		if nil != err {
			return nil, err
		}
	}

	//

	// FIXME: this is because etcd loader is bad at handling pointers,
	// we need default values for this to work
	err = revip.Postprocess(
		c,
		InitPostprocessors...,
	)
	if err != nil {
		return nil, err
	}

	_, err = revip.Load(
		c,
		append(loaders, revip.FromEnviron(EnvironPrefix))...,
	)
	if err != nil {
		return nil, err
	}

	err = revip.Postprocess(
		c,
		postprocessors...,
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func Validate(c *Config) error {
	return revip.WithValidation()(c)
}
