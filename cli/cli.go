package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	watchdog "github.com/cloudflare/tableflip"
	revip "github.com/corpix/revip"
	spew "github.com/davecgh/go-spew/spew"
	cli "github.com/urfave/cli/v2"
	di "go.uber.org/dig"

	"git.backbone/corpix/unregistry/pkg/bus"
	"git.backbone/corpix/unregistry/pkg/config"
	"git.backbone/corpix/unregistry/pkg/errors"
	"git.backbone/corpix/unregistry/pkg/log"
	"git.backbone/corpix/unregistry/pkg/meta"
	"git.backbone/corpix/unregistry/pkg/registry"
	"git.backbone/corpix/unregistry/pkg/telemetry"
)

var (
	Stdout = os.Stdout
	Stderr = os.Stderr

	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "pid-file",
			Aliases: []string{"p"},
			EnvVars: []string{config.EnvironPrefix + "_PID_FILE"},
			Usage:   "path to pid file to report into",
			Value:   meta.Name + ".pid",
		},
		&cli.StringFlag{
			Name:    "log-level",
			Aliases: []string{"l"},
			Usage:   "logging level (debug, info, warn, error)",
		},
		&cli.StringSliceFlag{
			Name:    "config",
			Aliases: []string{"c"},
			EnvVars: []string{config.EnvironPrefix + "_CONFIG"},
			Usage:   "path to application configuration file/files (separate multiple files with comma)",
			Value:   cli.NewStringSlice("config.yml"),
		},

		//

		&cli.DurationFlag{
			Name:  "duration",
			Usage: "exit after duration",
		},
		&cli.BoolFlag{
			Name:  "profile",
			Usage: "write profile information for debugging (cpu.prof, heap.prof)",
		},
		&cli.BoolFlag{
			Name:  "trace",
			Usage: "write trace information for debugging (trace.prof)",
		},
	}
	Commands = []*cli.Command{
		{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Configuration tools",
			Subcommands: []*cli.Command{
				{
					Name:    "show-default",
					Aliases: []string{"sd"},
					Usage:   "Show default configuration",
					Action:  ConfigShowDefaultAction,
				},
				{
					Name:    "show",
					Aliases: []string{"s"},
					Usage:   "Show default configuration",
					Action:  ConfigShowAction,
				},
				{
					Name:    "validate",
					Aliases: []string{"v"},
					Usage:   "Validate configuration and exit",
					Action:  ConfigValidateAction,
				},
				{
					Name:      "push",
					Aliases:   []string{"p"},
					Usage:     "Push configuration to specified destination",
					Action:    ConfigPushAction,
					ArgsUsage: "<destination>[,...]",
				},
			},
		},
	}

	c *di.Container
)

func Before(ctx *cli.Context) error {
	var err error

	c = di.New()

	//

	err = c.Provide(func() *cli.Context { return ctx })
	if err != nil {
		return err
	}

	err = c.Provide(func() *spew.ConfigState {
		return &spew.ConfigState{
			DisableMethods:          false,
			DisableCapacities:       true,
			DisablePointerAddresses: true,
			Indent:                  "  ",
			SortKeys:                true,
			SpewKeys:                false,
		}
	})
	if err != nil {
		return err
	}

	err = c.Provide(func() *json.Encoder {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc
	})
	if err != nil {
		return err
	}

	err = c.Provide(func(ctx *cli.Context) (*config.Config, error) {
		c, err := config.Load(ctx.StringSlice("config"))
		if err != nil {
			return nil, err
		}

		return c, nil
	})
	if err != nil {
		return err
	}

	err = c.Provide(func(ctx *cli.Context, c *config.Config) (log.Logger, error) {
		lc := *c.Log
		level := ctx.String("log-level")
		if level != "" {
			lc.Level = level
		}

		return log.Create(lc)
	})
	if err != nil {
		return err
	}

	err = c.Provide(func() *telemetry.Registry { return telemetry.DefaultRegistry })
	if err != nil {
		return err
	}

	//

	err = c.Provide(func(
		c *config.Config,
		l log.Logger,
		r *telemetry.Registry,
		w *watchdog.Upgrader,
		running *sync.WaitGroup,
		errc chan error,
	) (*telemetry.Server, error) {
		start := func(t *telemetry.Server) {
			errc <- errors.Wrap(
				t.ListenAndServe(),
				"failed while listen and serve telemetry server",
			)
		}

		finalize := func(t *telemetry.Server) {
			defer running.Done()
			<-w.Exit()

			err = t.Shutdown(context.Background())
			if err != nil {
				panic(errors.Wrap(err, "telemetry server shutdown failed"))
			}
		}

		if c.Telemetry.Enable {
			lr, err := w.Listen("tcp", c.Telemetry.Addr)
			if err != nil {
				return nil, err
			}
			t, err := telemetry.New(*c.Telemetry, l, r, lr)
			if err != nil {
				return nil, err
			}

			running.Add(1)

			go start(t)
			go finalize(t)

			return t, nil
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	//

	err = c.Provide(func(c *config.Config) (registry.Provider, error) {
		return registry.NewProvider(*c.Registry.Provider)
	})
	if err != nil {
		return err
	}

	err = c.Provide(func(
		c *config.Config,
		l log.Logger,
		tr *telemetry.Registry,
		p registry.Provider,
		w *watchdog.Upgrader,
		running *sync.WaitGroup,
		errc chan error,
	) (*registry.Server, error) {
		start := func(r *registry.Server) {
			errc <- errors.Wrap(
				r.ListenAndServe(),
				"failed while listen and serve registry server",
			)
		}

		finalize := func(r *registry.Server) {
			defer running.Done()
			<-w.Exit()

			err = r.Shutdown(context.Background())
			if err != nil {
				panic(errors.Wrap(err, "registry server shutdown failed"))
			}
		}

		lr, err := w.Listen("tcp", c.Registry.Server.Addr)
		if err != nil {
			return nil, err
		}
		r, err := registry.NewServer(*c.Registry.Server, l, tr, lr, p)
		if err != nil {
			return nil, err
		}

		running.Add(1)

		go start(r)
		go finalize(r)

		return r, nil
	})
	if err != nil {
		return err
	}

	//

	err = c.Provide(func(ctx *cli.Context, c *config.Config) (*watchdog.Upgrader, error) {
		return watchdog.New(watchdog.Options{
			UpgradeTimeout: c.ShutdownGraceTime,
			PIDFile:        ctx.String("pid-file"),
		})
	})
	if err != nil {
		return err
	}

	err = c.Provide(func() *sync.WaitGroup { return &sync.WaitGroup{} })
	if err != nil {
		return err
	}

	err = c.Provide(func() chan error { return make(chan error, 1) })
	if err != nil {
		return err
	}

	err = c.Provide(func() chan os.Signal {
		sig := make(chan os.Signal, 1)
		signal.Notify(
			sig,
			syscall.SIGQUIT,
			syscall.SIGTERM,
			syscall.SIGINT,
			syscall.SIGUSR1,
			syscall.SIGUSR2,
			syscall.SIGHUP,
		)
		return sig
	})
	if err != nil {
		return err
	}

	//

	duration := ctx.Duration("duration")
	if duration == 0 {
		err = c.Provide(func(ctx *cli.Context) context.Context {
			return context.Background()
		})
	} else {
		err = c.Provide(func(ctx *cli.Context) context.Context {
			c, cancel := context.WithTimeout(context.Background(), duration)
			go func() {
				<-c.Done()
				cancel()
			}()
			return c
		})
	}
	if err != nil {
		return err
	}

	//

	if ctx.Bool("profile") {
		err = c.Invoke(writeProfile)
		if err != nil {
			return err
		}
	}

	if ctx.Bool("trace") {
		err = c.Invoke(writeTrace)
		if err != nil {
			return err
		}
	}

	return nil
}

//

func ConfigShowDefaultAction(ctx *cli.Context) error {
	c, err := config.Default()
	if err != nil {
		return err
	}

	write := revip.ToWriter(os.Stdout, config.Marshaler)

	return write(c)
}

func ConfigShowAction(ctx *cli.Context) error {
	return c.Invoke(func(c *config.Config) error {
		write := revip.ToWriter(os.Stdout, config.Marshaler)

		return write(c)
	})
}

func ConfigValidateAction(ctx *cli.Context) error {
	return c.Invoke(func(l log.Logger) error {
		configs := ctx.StringSlice("config")
		c, err := config.Load(
			configs,
			config.InitPostprocessors...,
		)
		if err != nil {
			return err
		}

		err = config.Validate(c)
		if err != nil {
			return err
		}

		l.Info().
			Strs("configs", configs).
			Msg("configuration validation is ok")

		return nil
	})
}

func ConfigPushAction(ctx *cli.Context) error {
	return c.Invoke(func(l log.Logger) error {
		configs := ctx.StringSlice("config")
		c, err := config.Load(
			configs,
			config.LocalPostprocessors...,
		)
		if err != nil {
			return err
		}

		args := ctx.Args().Slice()
		if len(args) < 1 {
			return errors.New("subcommand requires an argument, example: ./config.out.yml")
		}

		destinations := args
		for _, destination := range destinations {
			push, err := revip.ToURL(destination, config.Marshaler)
			if err != nil {
				return err
			}

			err = push(c)
			if err != nil {
				return err
			}
		}

		l.Info().
			Strs("configs", configs).
			Strs("destinations", destinations).
			Msg("configuration pushed")

		return nil
	})
}

//

func RootAction(ctx *cli.Context) error {
	components := c.String()
	_ = c.Invoke(func(l log.Logger) {
		l.Trace().Msgf(
			"component graph: %s",
			strings.TrimSpace(components),
		)
	})

	return c.Invoke(func(
		ctx context.Context,
		cfg *config.Config,
		w *watchdog.Upgrader,
		l log.Logger,
		t *telemetry.Server,
		r *registry.Server,
		running *sync.WaitGroup,
		errc chan error,
		sig chan os.Signal,
	) error {
		l.Info().Msg("running")

		err := w.Ready()
		if err != nil {
			return err
		}

		//

	loop:
		for {
			select {
			case <-w.Exit():
				break loop
			case <-ctx.Done():
				w.Stop()
				break loop

			case err := <-errc:
				if err != nil {
					return err
				}
			case si := <-sig:
				l.Info().Str("signal", si.String()).Msg("received signal")
				switch si {
				case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					w.Stop()
				case syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP:
					err = w.Upgrade()
					if err != nil {
						return err
					}
				}
			case <-bus.Config:
				err = w.Upgrade()
				if err != nil {
					return err
				}
			}
		}

		//

		defer os.Exit(0)
		l.Info().Msg("shutdown watchdog")

		time.AfterFunc(cfg.ShutdownGraceTime, func() {
			l.Warn().
				Dur("graceTime", cfg.ShutdownGraceTime).
				Msg("graceful shutdown timed out")
			os.Exit(1)
		})

		running.Wait() // wait for other running components to finish

		return nil
	})
}

//

func NewApp() *cli.App {
	app := &cli.App{}

	app.Before = Before
	app.Flags = Flags
	app.Action = RootAction
	app.Commands = Commands
	app.Version = meta.Version

	return app
}

func Run() {
	err := NewApp().Run(os.Args)
	if err != nil {
		errors.Fatal(errors.Wrap(
			err, fmt.Sprintf(
				"pid: %d, ppid: %d",
				os.Getpid(), os.Getppid(),
			),
		))
	}
}
