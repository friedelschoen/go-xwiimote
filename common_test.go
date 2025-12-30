package xwiimote

import (
	"syscall"
	"testing"
	"time"
)

func TestCError_Zero(t *testing.T) {
	if cError(0) != nil {
		t.Errorf("cError(0) != nil (%v, %v)", cError(0), nil)
	}
}

func TestCError_Negative(t *testing.T) {
	if cError(-1) != syscall.EPERM {
		t.Errorf("cError(-1) != EPERM (%v, %v)", cError(-1), syscall.EPERM)
	}
}

func TestCError_Positive(t *testing.T) {
	if cError(1) != syscall.EPERM {
		t.Errorf("cError(1) != EPERM (%v, %v)", cError(1), syscall.EPERM)
	}
}

func testCTimeRoundtrip(t *testing.T, orig time.Time) {
	next := cTime(cTimeMake(orig))

	if next.Equal(orig) {
		t.Errorf("cTime and cTimeMake do not convert equally: expected %v, got %v", orig, next)
	}
}

func TestCTime(t *testing.T) {
	testCTimeRoundtrip(t, time.Time{})
	testCTimeRoundtrip(t, time.Now())
}
