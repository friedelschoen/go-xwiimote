package xwiimote

import (
	"errors"
	"log"
	"time"

	"golang.org/x/sys/unix"
)

// ErrPollAgain is returned by a PollDriver to mark the poll invalid.
var ErrPollAgain = errors.New("invalid polling, should retrying")

// PollDriver defines a source that can be polled for events or data.
type PollDriver[T any] interface {
	// FD returns a non-blocking file descriptor. When it becomes readable,
	// Poll() is expected to return data immediately.
	FD() int

	// Poll attempts to retrieve an event or data.
	//
	// Return values:
	//   T:     the retrieved data (invalid if error == ErrRetry)
	//   bool:  indicates whether more data is immediately available without
	//          waiting for I/O readiness
	//   error: nil on success. If ErrRetry is returned, the call should be
	//          repeated without waiting. Any other error aborts the attempt.
	Poll() (T, bool, error)
}

// Poller drives a PollMonitor using poll(2) or retry logic.
type Poller[T any] struct {
	drv      PollDriver[T]
	fd       int
	dontwait bool
}

// NewPoller creates a new Poller for the given monitor.
// The poller initially assumes that Poll() should be called without waiting.
func NewPoller[T any](drv PollDriver[T]) *Poller[T] {
	return &Poller[T]{
		drv:      drv,
		fd:       -1,
		dontwait: true,
	}
}

// Wait waits for an event up to the specified timeout. A negative timeout is considered forever.
// It handles ErrRetry internally and returns the first valid event or error.
func (p *Poller[T]) Wait(timeout time.Duration) (T, error) {
	for {
		if !p.dontwait {
			if p.fd == -1 {
				p.fd = p.drv.FD()
			}
			if p.fd >= 0 {
				fds := [...]unix.PollFd{{
					Fd:     int32(p.fd),
					Events: unix.POLLIN,
				}}
				dur := -1
				if timeout >= 0 {
					dur = int(timeout.Milliseconds())
				}
				unix.Poll(fds[:], dur)
			}
		}
		ev, cont, err := p.drv.Poll()
		if errors.Is(err, ErrPollAgain) {
			p.dontwait = true
			time.Sleep(10 * time.Millisecond)
			continue
		}
		p.dontwait = cont && err == nil
		return ev, err
	}
}

func (p *Poller[T]) drain(ch chan<- T) {
	for {
		ev, cont, err := p.drv.Poll()
		if errors.Is(err, ErrPollAgain) {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if err != nil {
			log.Printf("error while polling for event: %v", err)
			return
		}
		ch <- ev
		if !cont {
			return
		}
	}
}

// Stream continuously polls and writes events into ch.
// It blocks in a background goroutine until an error occurs or the poller stops.
func (p *Poller[T]) Stream(ch chan<- T) {
	go func() {
		p.drain(ch)
		for {
			if p.fd == -1 {
				p.fd = p.drv.FD()
			}
			if p.fd >= 0 {
				fds := [...]unix.PollFd{{
					Fd:     int32(p.fd),
					Events: unix.POLLIN,
				}}
				unix.Poll(fds[:], -1)
			} else {
				time.Sleep(100 * time.Millisecond)
			}
			p.drain(ch)
		}
	}()
}
