package xwiimote

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type pollStep[T any] struct {
	ev   T
	cont bool
	err  error
}

type fakeDriver[T any] struct {
	mu        sync.Mutex
	fd        int
	fdCalls   int
	pollCalls int
	steps     []pollStep[T]
}

func (d *fakeDriver[T]) FD() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.fdCalls++
	return d.fd
}

func (d *fakeDriver[T]) Poll() (T, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var zero T
	d.pollCalls++

	if len(d.steps) == 0 {
		return zero, false, errors.New("no more steps")
	}
	s := d.steps[0]
	d.steps = d.steps[1:]
	return s.ev, s.cont, s.err
}

func TestPollerWait_RetriesOnErrPollAgain(t *testing.T) {
	d := &fakeDriver[int]{
		fd: -1, // voorkomt unix.Poll pad als dontwait false zou worden
		steps: []pollStep[int]{
			{ev: 0, cont: false, err: ErrPollAgain},
			{ev: 42, cont: false, err: nil},
		},
	}
	p := newPoller(d)

	start := time.Now()
	ev, err := p.Wait(0)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if ev != 42 {
		t.Fatalf("expected ev=42, got %v", ev)
	}

	// Zorg dat we echt gere-try'd hebben.
	if d.pollCalls < 2 {
		t.Fatalf("expected >=2 Poll calls, got %d", d.pollCalls)
	}

	// De code slaapt 10ms bij ErrPollAgain; wees niet superstrak, maar check dat het niet instant was.
	if time.Since(start) < 5*time.Millisecond {
		t.Fatalf("expected a small delay due to retry sleep; got %v", time.Since(start))
	}
}

func TestPollerWait_ContKeepsDontWaitTrue(t *testing.T) {
	d := &fakeDriver[int]{
		fd: -1,
		steps: []pollStep[int]{
			{ev: 1, cont: true, err: nil}, // cont=true => dontwait blijft true
			{ev: 2, cont: false, err: nil},
		},
	}
	p := newPoller(d)

	ev, err := p.Wait(0)
	if err != nil || ev != 1 {
		t.Fatalf("first Wait: expected (1,nil), got (%v,%v)", ev, err)
	}

	// Als dontwait true blijft, mag FD() niet aangeroepen worden.
	if d.fdCalls != 0 {
		t.Fatalf("expected FD() not called yet, got %d", d.fdCalls)
	}

	ev, err = p.Wait(0)
	if err != nil || ev != 2 {
		t.Fatalf("second Wait: expected (2,nil), got (%v,%v)", ev, err)
	}
	if d.fdCalls != 0 {
		t.Fatalf("expected FD() still not called, got %d", d.fdCalls)
	}
}

func TestPollerWait_CallsFDOnlyWhenDontWaitFalse(t *testing.T) {
	d := &fakeDriver[int]{
		fd: -1, // Cruciaal: zo ga je niet via unix.Poll, maar je test wel dat FD() opgevraagd wordt.
		steps: []pollStep[int]{
			{ev: 7, cont: false, err: nil},
		},
	}
	p := newPoller(d)

	// Forceer pad: !dontwait => hij gaat FD() halen.
	p.wait = true

	ev, err := p.Wait(0)
	if err != nil || ev != 7 {
		t.Fatalf("expected (7,nil), got (%v,%v)", ev, err)
	}

	if d.fdCalls != 1 {
		t.Fatalf("expected FD() called exactly once, got %d", d.fdCalls)
	}

	// cont=false => dontwait moet false blijven na return.
	if !p.wait {
		t.Fatalf("expected dontwait=false after cont=false, got true")
	}
}
