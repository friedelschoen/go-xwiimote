package xwiimote

//go:generate morestringer -type Led,Key,KeyState,MonitorType -output stringer.go

// #include <string.h>
// #include <stdlib.h>
// #include <linux/input.h>
// #include <errno.h>
//
// unsigned int eviocgname(size_t sz) { return EVIOCGNAME(sz); }
import "C"
import (
	"syscall"
	"time"
)

// cError takes an integer error code.
//
// code == 0 -> nil
// code < 0  -> syscall.Errno(-code)
// code > 0  -> syscall.Errno(code)
func cError(ret C.int) error {
	if ret == 0 {
		return nil
	}
	if ret < 0 {
		ret = -ret
	}
	return syscall.Errno(ret)
}

func cTimeMake(orig time.Time) C.struct_timeval {
	var t C.struct_timeval
	t.tv_sec = C.time_t(orig.Unix())
	t.tv_usec = C.time_t(orig.UnixMicro())
	return t
}

// cTime takes an C timeval and converts it to time.Time
func cTime(t C.struct_timeval) time.Time {
	return time.Unix(int64(t.tv_sec), int64(t.tv_usec)*1000)
}

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if err == 0 {
		return nil
	}
	return err
}
