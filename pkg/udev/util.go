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

// enumeratorIterator creates an iterator over an udev_list_entry, init() should return the initial entry.
func enumeratorIterator(ctx *udevContext, init func() *C.struct_udev_list_entry) iter.Seq[*C.struct_udev_list_entry] {
	return sequences.Unfold(init, func(l *C.struct_udev_list_entry) (*C.struct_udev_list_entry, bool) {
		ctx.lock()
		defer ctx.unlock()
		next := C.udev_list_entry_get_next(l)
		return next, next != nil
	})
}

// enumeratorIterator creates an iterator over names of udev_list_entry, init() should return the initial entry.
func enumerateName(ctx *udevContext, init func() *C.struct_udev_list_entry) iter.Seq[string] {
	return sequences.Map(enumeratorIterator(ctx, init), func(l *C.struct_udev_list_entry) string {
		ctx.lock()
		defer ctx.unlock()
		return C.GoString(C.udev_list_entry_get_name(l))
	})
}

// enumeratorIterator creates an iterator over names and values of udev_list_entry, init() should return the initial entry.
func enumerateNameValue(ctx *udevContext, init func() *C.struct_udev_list_entry) iter.Seq2[string, string] {
	return sequences.Map12(enumeratorIterator(ctx, init), func(l *C.struct_udev_list_entry) (string, string) {
		ctx.lock()
		defer ctx.unlock()
		return C.GoString(C.udev_list_entry_get_name(l)), C.GoString(C.udev_list_entry_get_value(l))
	})
}
