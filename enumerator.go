package xwiimote

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
import "C"
import (
	"runtime"
)

// Enumerator object
//
// Each object describes a separate monitor. A single monitor must not be
// used from multiple threads without locking. Different monitors are
// independent of each other and can be used simultaneously.
type Enumerator struct {
	cptr *C.struct_xwii_monitor
}

// Create a new monitor
//
// Creates a new monitor and returns a pointer to the opaque object. NULL is
// returned on failure.
//
// @param[in] poll True if this monitor should watch for hotplug events
// @param[in] direct True if kernel uevents should be used instead of udevd
//
// A monitor always provides all devices that are available on a system. If
// @p poll is true, the monitor also sets up a system-monitor to watch the
// system for new hotplug events so new devices can be detected.
//
// A new monitor always has a ref-count of 1.
func NewEnumerator(direct bool) *Enumerator {
	enum := new(Enumerator)
	enum.cptr = C.xwii_monitor_new(false, C.bool(direct))

	runtime.SetFinalizer(enum, func(e *Enumerator) {
		e.Free()
	})
	return enum
}

func (enum *Enumerator) Free() {
	if enum.cptr == nil {
		return
	}
	runtime.SetFinalizer(enum, nil)
	C.xwii_monitor_unref(enum.cptr)
	enum.cptr = nil
}

// Read incoming events
//
// @param[in] monitor A valid monitor object
//
// This returns a single device-name on each call. A device-name is actually
// an absolute sysfs path to the device's root-node. This is normally a path
// to /sys/bus/hid/devices/[dev]/. You can use this path to create a new
// struct xwii_iface object.
//
// After a monitor was created, this function returns all currently available
// devices. After all devices have been returned, this function returns an empty string.
func (enum *Enumerator) Next() string {
	path := C.xwii_monitor_poll(enum.cptr)
	return cStringCopy(path)
}
