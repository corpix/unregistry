package registry

import (
	"context"
	"net"
	"regexp"

	echomw "github.com/labstack/echo/v4/middleware"

	"git.backbone/corpix/unregistry/pkg/log"
	"git.backbone/corpix/unregistry/pkg/server"
	"git.backbone/corpix/unregistry/pkg/telemetry"
)

const (
	Prefix    = "/v2"
	Namespace = Prefix + "/"
)

var (
	manifestRegex = regexp.MustCompile(`^` + Prefix + `/([\w|\-|\.|\_|\/]+)/manifests/([^/]+)$`)
	blobRegex     = regexp.MustCompile(`^` + Prefix + `/([\w|\-|\.|\_|\/]+)/blobs/([^/]+)$`)
)

//

type ServerConfig struct {
	Addr string         `yaml:"addr"`
	HTTP *server.Config `yaml:"http"`
}

func (c *ServerConfig) Default() {
loop:
	for {
		switch {
		case c.Addr == "":
			c.Addr = "127.0.0.1:5000"
		case c.HTTP == nil:
			c.HTTP = &server.Config{}
		default:
			break loop
		}
	}
}

//

type (
	Listener net.Listener
	Server   struct {
		config   ServerConfig
		log      log.Logger
		srv      *server.Server
		provider Provider
	}
)

func (s *Server) ListenAndServe() error {
	err := s.srv.StartServer(
		server.NewHTTPServer(
			s.config.Addr,
			server.HTTPTimeoutOption(*s.config.HTTP.Timeout),
		),
	)
	if err == server.ErrServerClosed {
		s.log.
			Warn().
			Str("addr", s.config.Addr).
			Msg("server shutdown")
		return nil
	}

	return err
}

func (s *Server) Mount(srv *server.Server) {
	srv.Router(Prefix, func(next server.HandlerFunc) server.HandlerFunc {
		return func(ctx server.Context) error {
			uri := ctx.Request().RequestURI

			if uri == Namespace {
				return ctx.NoContent(server.StatusOK)
			}

			if match := manifestRegex.FindStringSubmatch(uri); len(match) == 3 {
				name, reference := match[1], match[2]
				stream, err := s.provider.GetManifest(name, reference)
				if err != nil {
					return err
				}
				defer stream.Close()

				return ctx.Stream(server.StatusOK, ManifestMediaType, stream)
			}

			if match := blobRegex.FindStringSubmatch(uri); len(match) == 3 {
				name, reference := match[1], match[2]
				stream, err := s.provider.GetBlob(name, reference)
				if err != nil {
					return err
				}
				defer stream.Close()

				return ctx.Stream(server.StatusOK, LayerMediaType, stream)
			}

			return next(ctx)
		}
	})
}

func (s *Server) Close() error {
	err := s.srv.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func NewServer(c ServerConfig, l log.Logger, r *telemetry.Registry, lr Listener, p Provider) (*Server, error) {
	var addr string

	if lr != nil {
		addr = lr.Addr().String()
	} else {
		addr = c.Addr
	}

	l = l.With().Str("component", Subsystem).Str("listener", addr).Logger()

	e, err := server.New(*c.HTTP, Subsystem, l, r)
	if err != nil {
		return nil, err
	}
	e.Listener = lr
	e.Use(echomw.BodyLimit("0"))

	s := &Server{
		config:   c,
		log:      l,
		srv:      e,
		provider: p,
	}

	s.Mount(s.srv)

	return s, nil
}
