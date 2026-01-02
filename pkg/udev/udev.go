// Package udev provides a cgo wrapper around the libudev C library
package udev

// #cgo pkg-config: libudev
// #include <libudev.h>
import "C"
import (
	"runtime"
	"sync"
)

// udevContext is an opaque struct wraping a udev library context
type udevContext struct {
	// A pointer to the C struct udev context
	udevPtr *C.struct_udev
	// Mutex for thread sync as libudev is not thread safe when called with the same struct udev
	m sync.Mutex
}

// Lock locks a udev context
func (u *udevContext) lock() {
	u.m.Lock()
}

// Unlock unlocks a udev context
func (u *udevContext) unlock() {
	u.m.Unlock()
}

// newDevice is a private helper function and returns a pointer to a new device.
// The device is also added t the devices map in the udev context.
// The agrument ptr is a pointer to the underlying C udev_device structure.
// The function returns nil if the pointer passed is NULL.
func newDevice() (d *Device) {
	// Create a new device object
	d = &Device{}
	d.udevPtr = C.udev_new()
	runtime.SetFinalizer(d, deviceUnref)
	// Return the device object
	return
}

// newMonitor is a private helper function and returns a pointer to a new monitor.
// The monitor is also added t the monitors map in the udev context.
// The agrument ptr is a pointer to the underlying C udev_monitor structure.
// The function returns nil if the pointer passed is NULL.
func newMonitor() (m *Monitor) {
	// Create a new device object
	m = &Monitor{}
	m.udevPtr = C.udev_new()
	runtime.SetFinalizer(m, monitorUnref)
	// Return the device object
	return
}

func newEnumerate() (e *Enumerate) {
	e = &Enumerate{}
	e.udevPtr = C.udev_new()
	runtime.SetFinalizer(e, enumerateUnref)
	// Return the device object
	return
}

// NewDeviceFromSyspath returns a pointer to a new device identified by its syspath, and nil on error
// The device is identified by the syspath argument
func NewDeviceFromSyspath(syspath string) *Device {
	d := newDevice()
	// Lock the udev context
	d.lock()
	defer d.unlock()
	// Convert Go strings to C strings for passing
	s := C.CString(syspath)
	defer freeCharPtr(s)
	// Return a new device
	d.ptr = C.udev_device_new_from_syspath(d.udevPtr, s)
	return d
}

// NewDeviceFromDevnum returns a pointer to a new device identified by its Devnum, and nil on error
// deviceType is 'c' for a character device and 'b' for a block device
func NewDeviceFromDevnum(deviceType uint8, n Devnum) *Device {
	d := newDevice()
	d.lock()
	defer d.unlock()
	d.ptr = C.udev_device_new_from_devnum(d.udevPtr, C.char(deviceType), n.d)
	return d
}

// NewDeviceFromSubsystemSysname returns a pointer to a new device identified by its subystem and sysname, and nil on error
func NewDeviceFromSubsystemSysname(subsystem, sysname string) *Device {
	d := newDevice()
	d.lock()
	defer d.unlock()
	ss, sn := C.CString(subsystem), C.CString(sysname)
	defer freeCharPtr(ss)
	defer freeCharPtr(sn)
	d.ptr = C.udev_device_new_from_subsystem_sysname(d.udevPtr, ss, sn)
	return d
}

// NewDeviceFromDeviceID returns a pointer to a new device identified by its device id, and nil on error
func NewDeviceFromDeviceID(id string) *Device {
	d := newDevice()
	d.lock()
	defer d.unlock()
	i := C.CString(id)
	defer freeCharPtr(i)
	d.ptr = C.udev_device_new_from_device_id(d.udevPtr, i)
	return d
}

// NewEnumerate returns a pointer to a new enumerate, and nil on error
func NewEnumerate() *Enumerate {
	e := newEnumerate()
	e.lock()
	defer e.unlock()
	e.ptr = C.udev_enumerate_new(e.udevPtr)
	return e
}

// NewMonitorFromNetlink returns a pointer to a new monitor listening to a NetLink socket, and nil on error
// The name argument is either "kernel" or "udev".
// When passing "kernel" the events are received before they are processed by udev.
// When passing "udev" the events are received after udev has processed the events and created device nodes.
// In most cases you will want to use "udev".
func NewMonitorFromNetlink(name string) *Monitor {
	m := newMonitor()
	m.lock()
	defer m.unlock()
	n := C.CString(name)
	defer freeCharPtr(n)
	m.ptr = C.udev_monitor_new_from_netlink(m.udevPtr, n)
	return m
}
