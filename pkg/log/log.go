package log

import (
	"io"
	"os"
	"syscall"

	console "github.com/mattn/go-isatty"
	"github.com/rs/zerolog"

	"git.backbone/corpix/unregistry/pkg/errors"
)

type (
	Level  = zerolog.Level
	Logger = zerolog.Logger
	Event  = zerolog.Event
)

const (
	Trace = zerolog.TraceLevel
	Debug = zerolog.DebugLevel
	Info  = zerolog.InfoLevel
	Warn  = zerolog.WarnLevel
	Error = zerolog.ErrorLevel
	Panic = zerolog.PanicLevel
	Fatal = zerolog.FatalLevel
)

const Subsystem = "log"

func Create(c Config) (Logger, error) {
	var (
		output = os.Stdout

		log   Logger
		level Level
		err   error
		w     io.Writer
	)

	if console.IsTerminal(output.Fd()) {
		w = zerolog.ConsoleWriter{Out: output}
	} else {
		w = output
	}

	level, err = zerolog.ParseLevel(c.Level)
	if err != nil {
		return log, errors.Wrap(err, "failed to parse logging level from config")
	}

	pgid, err := syscall.Getpgid(os.Getpid())
	if err != nil {
		panic(err)
	}

	log = zerolog.New(w).With().
		Int("pid", os.Getpid()).
		Int("ppid", os.Getppid()).
		Int("pgid", pgid).
		Timestamp().Logger().
		Level(level)

	return log, nil
}
