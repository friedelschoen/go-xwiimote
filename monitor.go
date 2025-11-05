package xwiimote

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
import "C"
import (
	"runtime"
)

/**
 * Monitor object
 *
 * Each object describes a separate monitor. A single monitor must not be
 * used from multiple threads without locking. Different monitors are
 * independent of each other and can be used simultaneously.
 */
type Monitor struct {
	cptr *C.struct_xwii_monitor
}

/**
 * Create a new monitor
 *
 * Creates a new monitor and returns a pointer to the opaque object. NULL is
 * returned on failure.
 *
 * @param[in] poll True if this monitor should watch for hotplug events
 * @param[in] direct True if kernel uevents should be used instead of udevd
 *
 * A monitor always provides all devices that are available on a system. If
 * @p poll is true, the monitor also sets up a system-monitor to watch the
 * system for new hotplug events so new devices can be detected.
 *
 * A new monitor always has a ref-count of 1.
 */
func NewMonitor(hotplug, direct bool) *Monitor {
	var mon Monitor
	mon.cptr = C.xwii_monitor_new(C.bool(hotplug), C.bool(direct))

	runtime.SetFinalizer(&mon, func(i *Device) {
		i.Free()
	})
	return &mon
}

func (mon *Monitor) Free() {
	if mon.cptr == nil {
		return
	}
	runtime.SetFinalizer(&mon, nil)
	C.xwii_monitor_unref(mon.cptr)
	mon.cptr = nil
}

/**
 * Return internal fd
 *
 * @param[in] monitor A valid monitor object
 * @param[in] blocking True to set the monitor in blocking mode
 *
 * Returns the file-descriptor used by this monitor. If @p blocking is true,
 * the FD is set into blocking mode. If false, it is set into non-blocking mode.
 * Only one file-descriptor exists, that is, this function always returns the
 * same descriptor.
 *
 * This returns -1 if this monitor was not created with a hotplug-monitor. So
 * you need this function only if you want to watch the system for hotplug
 * events. Whenever this descriptor is readable, you should call
 * xwii_monitor_poll() to read new incoming events.
 */
func (mon *Monitor) GetFD(blocking bool) (int, bool) {
	ret := C.xwii_monitor_get_fd(mon.cptr, C.bool(blocking))
	return int(ret), ret != -1
}

/**
 * Read incoming events
 *
 * @param[in] monitor A valid monitor object
 *
 * This returns a single device-name on each call. A device-name is actually
 * an absolute sysfs path to the device's root-node. This is normally a path
 * to /sys/bus/hid/devices/[dev]/. You can use this path to create a new
 * struct xwii_iface object.
 *
 * After a monitor was created, this function returns all currently available
 * devices. After all devices have been returned, this function returns NULL
 * _once_. After that, this function polls the monitor for hotplug events and
 * returns hotplugged devices, if the monitor was opened to watch the system for
 * hotplug events.
 * Use xwii_monitor_get_fd() to get notified when a new event is available. If
 * the fd is in non-blocking mode, this function never blocks but returns NULL
 * if no new event is available.
 *
 * The returned string must be freed with free() by the caller.
 */
func (mon *Monitor) Poll() string {
	path := C.xwii_monitor_poll(mon.cptr)
	return cStringCopy(path)
}
