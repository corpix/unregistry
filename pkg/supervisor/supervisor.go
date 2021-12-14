package supervisor

import (
	"time"

	"git.backbone/corpix/unregistry/pkg/errors"
	"git.backbone/corpix/unregistry/pkg/log"
	"git.backbone/corpix/unregistry/pkg/sync"
)

const subsystem = "supervisor"

type Thunk = func() error

type Strategy interface {
	Handle(string, error) bool
}

type Supervisor struct {
	name      string
	strategy  Strategy
	semaphore *sync.Semaphore
}

func (s Supervisor) Supervise(f Thunk) error {
	if !s.semaphore.TryWait() {
		return errors.New("failed to acquire semaphore, this means other task is running on this supervisor")
	}
	defer s.semaphore.Post() // nolint: errcheck

	var err error
	for {
		err = f()
		if err != nil && !s.strategy.Handle(s.name, err) {
			break
		}
	}
	return err
}

func New(n string, s Strategy) Supervisor { return Supervisor{n, s, sync.NewSemaphore(1, 1)} }

//

type DelayRestartStrategy struct {
	log   log.Logger
	delay time.Duration
}

func (s DelayRestartStrategy) Handle(name string, err error) bool {
	s.log.Error().Err(err).
		Str("component", subsystem).
		Str("supervisor", name).
		Msgf("task failed, restarting in %s", s.delay)

	time.Sleep(s.delay)
	return true
}

func NewDelayRestartStrategy(l log.Logger, d time.Duration) DelayRestartStrategy {
	return DelayRestartStrategy{l, d}
}
