package udev

// #cgo pkg-config: libudev
// #include <libudev.h>
// #include <stdlib.h>
import "C"

import (
	"iter"
	"unsafe"

	"github.com/friedelschoen/go-xwiimote/pkg/udev/sequences"
)

func freeCharPtr(s *C.char) {
	C.free(unsafe.Pointer(s))
}

func enumeratorIterator(locker *udevContext, init func() *C.struct_udev_list_entry) iter.Seq[*C.struct_udev_list_entry] {
	return sequences.Unfold(init, func(l *C.struct_udev_list_entry) (*C.struct_udev_list_entry, bool) {
		locker.lock()
		defer locker.unlock()
		next := C.udev_list_entry_get_next(l)
		return next, next != nil
	})
}

func enumerateName(locker *udevContext, init func() *C.struct_udev_list_entry) iter.Seq[string] {
	return sequences.Map(enumeratorIterator(locker, init), func(l *C.struct_udev_list_entry) string {
		locker.lock()
		defer locker.unlock()
		return C.GoString(C.udev_list_entry_get_name(l))
	})
}

func enumerateNameValue(locker *udevContext, init func() *C.struct_udev_list_entry) iter.Seq2[string, string] {
	return sequences.Map12(enumeratorIterator(locker, init), func(l *C.struct_udev_list_entry) (string, string) {
		locker.lock()
		defer locker.unlock()
		return C.GoString(C.udev_list_entry_get_name(l)), C.GoString(C.udev_list_entry_get_value(l))
	})
}
