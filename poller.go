package xwiimote

import (
	"errors"
	"log"
	"time"

	"golang.org/x/sys/unix"
)

// ErrPollAgain is returned by a PollDriver to mark the poll invalid.
var ErrPollAgain = errors.New("invalid polling, should retrying")

// pollDriver defines a source that can be polled for events or data.
type pollDriver[T any] interface {
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

// poller drives a PollMonitor using poll(2) or retry logic.
type poller[T any] struct {
	drv      pollDriver[T]
	fd       int
	dontwait bool
}

// newPoller creates a new Poller for the given monitor.
// The poller initially assumes that Poll() should be called without waiting.
func newPoller[T any](drv pollDriver[T]) poller[T] {
	return poller[T]{
		drv:      drv,
		fd:       -1,
		dontwait: true,
	}
}

// Wait waits for an event up to the specified timeout. A negative timeout is considered forever.
// It handles ErrRetry internally and returns the first valid event or error.
func (p *poller[T]) Wait(timeout time.Duration) (T, error) {
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

func (p *poller[T]) drain(yield func(T)) {
	for {
		ev, _, err := p.drv.Poll()
		if errors.Is(err, ErrPollAgain) {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if err != nil {
			log.Printf("error while polling for event: %v", err)
			continue
		}
		yield(ev)
	}
}

// Handle continuously polls and calls `yield` with new events.
// It blocks forever and should be used in a new goroutine.
func (p *poller[T]) Handle(yield func(T)) {
	p.drain(yield)
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
		p.drain(yield)
	}
}

// Stream continuously polls and writes events into ch. It is a wrapper for Handle.
// It blocks forever and should be used in a new goroutine.
//
//	p.Handle(func(ev T) { ch <- ev })
func (p *poller[T]) Stream(ch chan<- T) {
	p.Handle(func(ev T) {
		ch <- ev
	})
}
