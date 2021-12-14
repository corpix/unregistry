package sync

import (
	"sync"

	"git.backbone/corpix/unregistry/pkg/errors"
)

type (
	WaitGroup = sync.WaitGroup
	Mutex     = sync.Mutex
	RWMutex   = sync.RWMutex
)

func NewWaitGroup() *WaitGroup { return &WaitGroup{} }
func NewMutex() *Mutex         { return &Mutex{} }
func NewRWMutex() *RWMutex     { return &RWMutex{} }

//

type Semaphore struct {
	limit int
	ch    chan struct{}
}

func (s *Semaphore) Post() error {
	select {
	case s.ch <- struct{}{}:
	default:
		return errors.Errorf("hit semaphore limit %d", s.limit)
	}
	return nil
}
func (s *Semaphore) Wait() { <-s.ch }
func (s *Semaphore) TryWait() bool {
	select {
	case <-s.ch:
		return true
	default:
		return false
	}
}

func NewSemaphore(initial int, limit int) *Semaphore {
	ch := make(chan struct{}, limit)
	for initial != 0 {
		ch <- struct{}{}
		initial--
	}
	return &Semaphore{limit, ch}
}
