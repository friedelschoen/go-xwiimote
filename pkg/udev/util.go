package udev

/*
  #cgo LDFLAGS: -ludev
  #include <libudev.h>
  #include <linux/types.h>
  #include <stdlib.h>
	#include <linux/kdev_t.h>
*/
import "C"

import (
	"iter"
	"unsafe"
)

func freeCharPtr(s *C.char) {
	C.free(unsafe.Pointer(s))
}

func enumerateName(locker interface {
	lock()
	unlock()
}, init func() *C.struct_udev_list_entry) iter.Seq[string] {
	return func(yield func(string) bool) {
		var l *C.struct_udev_list_entry
		for {
			locker.lock()
			if l == nil {
				l = init()
			} else {
				l = C.udev_list_entry_get_next(l)
				if l == nil {
					return
				}
			}
			item := C.GoString(C.udev_list_entry_get_name(l))
			locker.unlock()

			if !yield(item) {
				return
			}
		}
	}
}
