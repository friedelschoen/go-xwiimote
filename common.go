package xwiimote

// #include <string.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"time"
	"unsafe"
)

func cError(ret C.int) error {
	if ret == 0 {
		return nil
	}
	if ret < 0 {
		ret = -ret
	}
	cerr := C.strerror(ret)
	return errors.New(C.GoString(cerr))
}

func cTime(t C.struct_timeval) time.Time {
	return time.Unix(int64(t.tv_sec), int64(t.tv_usec))
}

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
