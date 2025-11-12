package xwiimote

// #include <string.h>
// #include <stdlib.h>
import "C"
import (
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

// cTime takes an C timeval and converts it to time.Time
func cTime(t C.struct_timeval) time.Time {
	return time.Unix(int64(t.tv_sec), int64(t.tv_usec))
}

// cStringCopy takes a NUL-terminated C-string and copyies it into a string and the cstr is freed afterwards.
// If the input is nil it returns an empty string.
func cStringCopy(cstr *C.char) string {
	if cstr == nil {
		return ""
	}
	size := C.strlen(cstr)
	result := make([]byte, size)
	C.memcpy(unsafe.Pointer(&result[0]), unsafe.Pointer(cstr), size)
	C.free(unsafe.Pointer(cstr))
	return string(result)
}
