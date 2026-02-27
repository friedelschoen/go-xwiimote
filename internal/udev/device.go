package udev

// #cgo pkg-config: libudev
// #include <libudev.h>
// #include <stdlib.h>
import "C"
import (
	"github.com/friedelschoen/go-wiimote"
)

// Device wraps a libudev device object
type Device struct {
	udevContext
	ptr *C.struct_udev_device
}

func deviceUnref(d *Device) {
	C.udev_device_unref(d.ptr)
	C.udev_unref(d.udevPtr)
}

// Parent returns the parent Device, or nil if the receiver has no parent Device
func (d *Device) Parent() wiimote.DeviceInfo {
	d.lock()
	defer d.unlock()
	ptr := C.udev_device_get_parent(d.ptr)

	pd := newDevice()
	pd.ptr = ptr
	return pd
}

// Subsystem returns the subsystem string of the udev device.
// The string does not contain any "/".
func (d *Device) Subsystem() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_subsystem(d.ptr))
}

// Sysname returns the sysname of the udev device (e.g. ttyS3, sda1...).
func (d *Device) Sysname() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_sysname(d.ptr))
}

// Syspath returns the sys path of the udev device.
// The path is an absolute path and starts with the sys mount point.
func (d *Device) Syspath() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_syspath(d.ptr))
}

// Devnode returns the device node file name belonging to the udev device.
// The path is an absolute path, and starts with the device directory.
func (d *Device) Devnode() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_devnode(d.ptr))
}

// Driver returns the driver for the receiver
func (d *Device) Driver() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_driver(d.ptr))
}

// Action returns the action for the event.
// This is only valid if the device was received through a monitor.
// Devices read from sys do not have an action string.
// Usual actions are: add, remove, change, online, offline.
func (d *Device) Action() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_action(d.ptr))
}

// SysattrValue retrieves the content of a sys attribute file, and returns an empty string if there is no sys attribute value.
// The retrieved value is cached in the device.
// Repeated calls will return the same value and not open the attribute again.
func (d *Device) SysattrValue(sysattr string) string {
	d.lock()
	defer d.unlock()
	s := C.CString(sysattr)
	defer freeCharPtr(s)
	return C.GoString(C.udev_device_get_sysattr_value(d.ptr, s))
}
