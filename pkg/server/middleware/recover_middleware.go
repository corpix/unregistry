package middleware

import (
	"fmt"
	"runtime"

	echo "github.com/labstack/echo/v4"

	"git.backbone/corpix/unregistry/pkg/log"
)

type (
	// RecoverConfig defines the config for Recover middleware.
	RecoverConfig struct {
		// Size of the stack to be printed.
		// Optional. Default value 4KB.
		StackSize int `yaml:"stack_size"`

		// DisableStackAll disables formatting stack traces of all other goroutines
		// into buffer after the trace for the current goroutine.
		// Optional. Default value false.
		DisableStackAll bool `yaml:"disable_stack_all"`

		// DisablePrintStack disables printing stack trace.
		// Optional. Default value as false.
		DisablePrintStack bool `yaml:"disable_print_stack"`
	}
)

var (
	// DefaultRecoverConfig is the default Recover middleware config.
	DefaultRecoverConfig = &RecoverConfig{
		StackSize:         4 << 10, // 4 KB
		DisableStackAll:   false,
		DisablePrintStack: false,
	}
)

// NewRecover returns a Recover middleware with config.
func NewRecover(c *RecoverConfig, l log.Logger) echo.MiddlewareFunc {
	if c == nil {
		c = DefaultRecoverConfig
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					stack := make([]byte, c.StackSize)
					length := runtime.Stack(stack, !c.DisableStackAll)
					evt := l.Error()
					if !c.DisablePrintStack {
						evt.Str("stack", string(stack[:length]))
					}
					evt.Err(err).Msg("panic recover")

					ctx.Error(err)
				}
			}()
			return next(ctx)
		}
	}
}
