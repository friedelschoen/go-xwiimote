package xwiimote

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
import "C"
import (
	"runtime"
)

// Enumerator describes a single one-time enumerator for xwiimote-devices.
// Enumerators are not thread-safe.
type Enumerator struct {
	cptr *C.struct_xwii_monitor
}

// NewEnumerator creates a new enumerator.
//
// A monitor always provides all devices that are available on a system.
//
// The object and underlying structure is freed automatically by default.
func NewEnumerator(typ MonitorType) *Enumerator {
	enum := new(Enumerator)
	enum.cptr = C.xwii_monitor_new(false, C.bool(typ))

	runtime.SetFinalizer(enum, func(e *Enumerator) {
		e.Free()
	})
	return enum
}

// Free unreferences the enumerator and frees the underlying structure.
// Calling Free is not mandatory and is done automatically by default.
func (enum *Enumerator) Free() {
	if enum.cptr == nil {
		return
	}
	runtime.SetFinalizer(enum, nil)
	C.xwii_monitor_unref(enum.cptr)
	enum.cptr = nil
}

// Next returns a single device-name on each call. A device-name is actually
// an absolute sysfs path to the device's root-node. This is normally a path
// to /sys/bus/hid/devices/[dev]/. You can use this path to create a new
// Device object. If the enumerator is exhausted an empty string is returned and
// no new elements will be provided.
func (enum *Enumerator) Next() string {
	path := C.xwii_monitor_poll(enum.cptr)
	return cStringCopy(path)
}
