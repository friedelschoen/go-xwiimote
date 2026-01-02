package udev

// #cgo pkg-config: libudev
// #include <libudev.h>
// #include <stdlib.h>
import "C"
import (
	"errors"

	"iter"
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
func (d *Device) Parent() *Device {
	d.lock()
	defer d.unlock()
	ptr := C.udev_device_get_parent(d.ptr)

	pd := newDevice()
	pd.ptr = ptr
	return pd
}

// ParentWithSubsystemDevtype returns the parent Device with the given subsystem and devtype,
// or nil if the receiver has no such parent device
func (d *Device) ParentWithSubsystemDevtype(subsystem, devtype string) *Device {
	d.lock()
	defer d.unlock()
	ss, dt := C.CString(subsystem), C.CString(devtype)
	defer freeCharPtr(ss)
	defer freeCharPtr(dt)
	ptr := C.udev_device_get_parent_with_subsystem_devtype(d.ptr, ss, dt)
	pd := newDevice()
	pd.ptr = ptr
	return pd
}

// Devpath returns the kernel devpath value of the udev device.
// The path does not contain the sys mount point, and starts with a '/'.
func (d *Device) Devpath() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_devpath(d.ptr))
}

// Subsystem returns the subsystem string of the udev device.
// The string does not contain any "/".
func (d *Device) Subsystem() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_subsystem(d.ptr))
}

// Devtype returns the devtype string of the udev device.
func (d *Device) Devtype() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_devtype(d.ptr))
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

// Sysnum returns the trailing number of of the device name
func (d *Device) Sysnum() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_sysnum(d.ptr))
}

// Devnode returns the device node file name belonging to the udev device.
// The path is an absolute path, and starts with the device directory.
func (d *Device) Devnode() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_devnode(d.ptr))
}

// IsInitialized checks if udev has already handled the device and has set up
// device node permissions and context, or has renamed a network device.
//
// This is only implemented for devices with a device node or network interfaces.
// All other devices return 1 here.
func (d *Device) IsInitialized() bool {
	d.lock()
	defer d.unlock()
	return C.udev_device_get_is_initialized(d.ptr) != 0
}

// Devlinks returns an Iterator over the device links pointing to the device file of the udev device.
func (d *Device) Devlinks() iter.Seq[string] {
	return enumerateName(&d.udevContext, func() *C.struct_udev_list_entry {
		d.lock()
		defer d.unlock()
		return C.udev_device_get_devlinks_list_entry(d.ptr)
	})

}

// Properties returns an Iterator over the key/value device properties of the udev device.
func (d *Device) Properties() iter.Seq2[string, string] {
	return enumerateNameValue(&d.udevContext, func() *C.struct_udev_list_entry {
		d.lock()
		defer d.unlock()
		return C.udev_device_get_properties_list_entry(d.ptr)
	})
}

// Tags returns an Iterator over the tags attached to the udev device.
func (d *Device) Tags() iter.Seq[string] {
	return enumerateName(&d.udevContext, func() *C.struct_udev_list_entry {
		d.lock()
		defer d.unlock()
		return C.udev_device_get_tags_list_entry(d.ptr)
	})
}

// Sysattrs returns an Iterator over the systems attributes of the udev device.
func (d *Device) Sysattrs() iter.Seq[string] {
	return enumerateName(&d.udevContext, func() *C.struct_udev_list_entry {
		d.lock()
		defer d.unlock()
		return C.udev_device_get_sysattr_list_entry(d.ptr)
	})
}

// PropertyValue retrieves the value of a device property
func (d *Device) PropertyValue(key string) string {
	d.lock()
	defer d.unlock()
	k := C.CString(key)
	defer freeCharPtr(k)
	return C.GoString(C.udev_device_get_property_value(d.ptr, k))
}

// Driver returns the driver for the receiver
func (d *Device) Driver() string {
	d.lock()
	defer d.unlock()
	return C.GoString(C.udev_device_get_driver(d.ptr))
}

// Devnum returns the device major/minor number.
func (d *Device) Devnum() Devnum {
	d.lock()
	defer d.unlock()
	return Devnum{C.udev_device_get_devnum(d.ptr)}
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

// Seqnum returns the sequence number of the event.
// This is only valid if the device was received through a monitor.
// Devices read from sys do not have a sequence number.
func (d *Device) Seqnum() uint64 {
	d.lock()
	defer d.unlock()
	return uint64(C.udev_device_get_seqnum(d.ptr))
}

// UsecSinceInitialized returns the number of microseconds passed since udev set up the device for the first time.
// This is only implemented for devices with need to store properties in the udev database.
// All other devices return 0 here.
func (d *Device) UsecSinceInitialized() uint64 {
	d.lock()
	defer d.unlock()
	return uint64(C.udev_device_get_usec_since_initialized(d.ptr))
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

// SetSysattrValue sets the content of a sys attribute file, and returns an error if this fails.
func (d *Device) SetSysattrValue(sysattr, value string) (err error) {
	d.lock()
	defer d.unlock()
	sa, val := C.CString(sysattr), C.CString(value)
	defer freeCharPtr(sa)
	defer freeCharPtr(val)
	if C.udev_device_set_sysattr_value(d.ptr, sa, val) < 0 {
		err = errors.New("udev: udev_device_set_sysattr_value failed")
	}
	return
}

// HasTag checks if the udev device has the tag specified
func (d *Device) HasTag(tag string) bool {
	d.lock()
	defer d.unlock()
	t := C.CString(tag)
	defer freeCharPtr(t)
	return C.udev_device_has_tag(d.ptr, t) != 0
}
