package common

import (
	"errors"
	"log"
	"syscall"
	"time"

	"github.com/friedelschoen/go-wiimote"
	"golang.org/x/sys/unix"
)

// ErrPollAgain is returned by a PollDriver to mark the poll invalid.
var ErrPollAgain = errors.New("invalid polling, should retrying")

// pollerDriver defines a source that can be polled for events or data.
type pollerDriver[T any] interface {
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
	drv  pollerDriver[T]
	fd   int
	wait bool
}

// Newpoller creates a new poller for the given monitor.
// The poller initially assumes that Poll() should be called without waiting.
func NewPoller[T any](drv pollerDriver[T]) wiimote.Poller[T] {
	return &poller[T]{
		drv: drv,
		fd:  -1,
	}
}

func (p *poller[T]) Poll() (T, bool, error) {
	return p.drv.Poll()
}

func (p *poller[T]) WaitReadable(timeout time.Duration) error {
	// Pak de fd vers; caching is fragiel als de driver fd’s kan vervangen.
	fd := p.drv.FD()
	p.fd = fd

	if fd < 0 {
		// Driver heeft (nog) geen pollbare fd
		return nil
	}

	fds := []unix.PollFd{{
		Fd:     int32(fd),
		Events: unix.POLLIN,
	}}

	dur := -1
	if timeout >= 0 {
		dur = int(timeout.Milliseconds())
	}

	for {
		n, err := unix.Poll(fds, dur)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			p.fd = -1
			return err
		}
		if n == 0 {
			// timeout
			return nil
		}
		break
	}

	re := fds[0].Revents
	if re&(unix.POLLNVAL) != 0 {
		// fd ongeldig
		p.fd = -1
		return unix.EBADF
	}
	if re&(unix.POLLERR|unix.POLLHUP) != 0 {
		// device hangup of error: laat caller beslissen wat te doen
		return unix.EIO
	}
	return nil
}

func (p *poller[T]) Wait(timeout time.Duration) (T, error) {
	for {
		if p.wait {
			if err := p.WaitReadable(timeout); err != nil {
				var zero T
				return zero, err
			}
		}
		ev, moredata, err := p.drv.Poll()
		p.wait = !moredata
		if errors.Is(err, ErrPollAgain) {
			time.Sleep(10 * time.Millisecond)
			continue
		}
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

func (p *poller[T]) Handle(yield func(T)) error {
	p.drain(yield)
	for {
		if p.wait {
			if err := p.WaitReadable(-1); err != nil {
				return err
			}
		}
		if p.fd < 0 {
			time.Sleep(100 * time.Millisecond)
		}
		p.drain(yield)
	}
}

func (p *poller[T]) Stream(ch chan<- T) {
	p.Handle(func(ev T) {
		ch <- ev
	})
}
