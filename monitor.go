package xwiimote

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
import "C"
import (
	"runtime"
)

// MonitorType describes how a monitor or enumerator should look for devices.
type MonitorType uint

const (
	// Monitor uses kernel uevents
	MonitorKernel MonitorType = 1
	// Monitor uses udevd
	MonitorUdev MonitorType = 0
)

// Monitor describes a monitor for xwiimote-devices. This includes currently available
// but also hot-plugged devices.
//
// Monitors are not thread-safe.
type Monitor struct {
	poller[string]
	cptr *C.struct_xwii_monitor
}

// NewMonitor creates a new monitor.
//
// A monitor always provides all devices that are available on a system
// and hot-plugged devices.
//
// You can use Poller[T] over a Monitor to efficiently wait for new devices.
//
// The object and underlying structure is freed automatically by default.
func NewMonitor(typ MonitorType) *Monitor {
	mon := new(Monitor)
	mon.poller = newPoller(mon)
	mon.cptr = C.xwii_monitor_new(true, C.bool(typ != 0))

	runtime.SetFinalizer(mon, func(m *Monitor) {
		m.Free()
	})
	return mon
}

// Free unreferences the monitor and frees the underlying structure.
// Calling Free is not mandatory and is done automatically by default.
func (mon *Monitor) Free() {
	if mon.cptr == nil {
		return
	}
	runtime.SetFinalizer(mon, nil)
	C.xwii_monitor_unref(mon.cptr)
	mon.cptr = nil
}

// FD returns the file-descriptor to notify readiness. The FD is non-blocking.
// Only one file-descriptor exists, that is, this function always returns the
// same descriptor.
func (mon *Monitor) FD() int {
	ret := C.xwii_monitor_get_fd(mon.cptr, false)
	return int(ret)
}

// Poll returns a single device-name on each call. A device-name is actually
// an absolute sysfs path to the device's root-node. This is normally a path
// to /sys/bus/hid/devices/[dev]/. You can use this path to create a new
// struct xwii_iface object.
//
// After a monitor was created, this function returns all currently available
// devices. After all devices have been returned. After that, this function polls the
// monitor for hotplug events and returns hotplugged devices,
// if the monitor was opened to watch the system for hotplug events.
//
// Use FD() to get notified when a new event is available.
func (mon *Monitor) Poll() (string, bool, error) {
	path := C.xwii_monitor_poll(mon.cptr)
	if path == nil {
		return "", false, ErrPollAgain
	}
	return cStringCopy(path), false, nil
}
