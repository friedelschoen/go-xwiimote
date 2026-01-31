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
	"bytes"
	"os"
	"syscall"
	"time"
	"unsafe"
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
	t.tv_usec = C.time_t(orig.UnixMicro() / 1_000_000)
	return t
}

// cTime takes an C timeval and converts it to time.Time
func cTime(t C.struct_timeval) time.Time {
	return time.Unix(int64(t.tv_sec), int64(t.tv_usec))
}

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if err == 0 {
		return nil
	}
	return err
}

func devname(fd *os.File) (string, error) {
	var buffer [256]byte
	if err := ioctl(fd.Fd(), uintptr(C.eviocgname(C.size_t(len(buffer)))), uintptr(unsafe.Pointer(&buffer[0]))); err != nil {
		return "", err
	}
	length := bytes.IndexByte(buffer[:], 0)
	return string(buffer[:length]), nil
}
