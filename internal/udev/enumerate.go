package udev

// #cgo pkg-config: libudev
// #include <libudev.h>
import "C"

import (
	"errors"

	"iter"

	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/internal/sequences"
)

// Enumerate is an opaque struct wrapping a udev enumerate object.
type Enumerate struct {
	udevContext
	ptr *C.struct_udev_enumerate
}

// Unref the Enumerate object
func enumerateUnref(e *Enumerate) {
	C.udev_enumerate_unref(e.ptr)
	C.udev_unref(e.udevPtr)
}

// AddMatchSubsystem adds a filter for a subsystem of the device to include in the list.
func (e *Enumerate) AddMatchSubsystem(subsystem string) (err error) {
	e.lock()
	defer e.unlock()
	s := C.CString(subsystem)
	defer freeCharPtr(s)
	if C.udev_enumerate_add_match_subsystem(e.ptr, s) != 0 {
		err = errors.New("udev: udev_enumerate_add_match_subsystem failed")
	}
	return
}

// AddNomatchSubsystem adds a filter for a subsystem of the device to exclude from the list.
func (e *Enumerate) AddNomatchSubsystem(subsystem string) (err error) {
	e.lock()
	defer e.unlock()
	s := C.CString(subsystem)
	defer freeCharPtr(s)
	if C.udev_enumerate_add_nomatch_subsystem(e.ptr, s) != 0 {
		err = errors.New("udev: udev_enumerate_add_nomatch_subsystem failed")
	}
	return
}

// AddMatchSysattr adds a filter for a sys attribute at the device to include in the list.
func (e *Enumerate) AddMatchSysattr(sysattr, value string) (err error) {
	e.lock()
	defer e.unlock()
	s, v := C.CString(sysattr), C.CString(value)
	defer freeCharPtr(s)
	defer freeCharPtr(v)
	if C.udev_enumerate_add_match_sysattr(e.ptr, s, v) != 0 {
		err = errors.New("udev: udev_enumerate_add_match_sysattr failed")
	}
	return
}

// AddNomatchSysattr adds a filter for a sys attribute at the device to exclude from the list.
func (e *Enumerate) AddNomatchSysattr(sysattr, value string) (err error) {
	e.lock()
	defer e.unlock()
	s, v := C.CString(sysattr), C.CString(value)
	defer freeCharPtr(s)
	defer freeCharPtr(v)
	if C.udev_enumerate_add_nomatch_sysattr(e.ptr, s, v) != 0 {
		err = errors.New("udev: udev_enumerate_add_nomatch_sysattr failed")
	}
	return
}

// AddMatchSysname adds a filter for the name of the device to include in the list.
func (e *Enumerate) AddMatchSysname(sysname string) (err error) {
	e.lock()
	defer e.unlock()
	s := C.CString(sysname)
	defer freeCharPtr(s)
	if C.udev_enumerate_add_match_sysname(e.ptr, s) != 0 {
		err = errors.New("udev: udev_enumerate_add_match_sysname failed")
	}
	return
}

// AddMatchParent adds a filter for a parent Device to include in the list.
func (e *Enumerate) AddMatchParent(parent wiimote.DeviceInfo) error {
	e.lock()
	defer e.unlock()

	parentdev, ok := parent.(*Device)
	if !ok {
		return errors.New("not a udev device")
	}
	if C.udev_enumerate_add_match_parent(e.ptr, parentdev.ptr) != 0 {
		return errors.New("udev: udev_enumerate_add_match_parent failed")
	}
	return nil
}

// AddSyspath adds a device to the list of enumerated devices, to retrieve it back sorted in dependency order.
func (e *Enumerate) AddSyspath(syspath string) (err error) {
	e.lock()
	defer e.unlock()
	s := C.CString(syspath)
	defer freeCharPtr(s)
	if C.udev_enumerate_add_syspath(e.ptr, s) != 0 {
		err = errors.New("udev: udev_enumerate_add_syspath failed")
	}
	return
}

// Devices returns an Iterator over the device syspaths matching the filter, sorted in dependency order.
// The Iterator is using the github.com/jkeiser/iter package.
// Values are returned as an interface{} and should be cast to string.
func (e *Enumerate) Devices() (it iter.Seq[wiimote.DeviceInfo], err error) {
	e.lock()
	defer e.unlock()
	if C.udev_enumerate_scan_devices(e.ptr) < 0 {
		err = errors.New("udev: udev_enumerate_scan_devices failed")
		return
	}

	names := enumerateName(&e.udevContext, func() *C.struct_udev_list_entry {
		e.lock()
		defer e.unlock()
		return C.udev_enumerate_get_list_entry(e.ptr)
	})
	return sequences.Map(names, func(path string) wiimote.DeviceInfo { return NewDeviceFromSyspath(path) }), nil
}

// Subsystems returns an Iterator over the subsystem syspaths matching the filter, sorted in dependency order.
// The Iterator is using the github.com/jkeiser/iter package.
// Values are returned as an interface{} and should be cast to string.
func (e *Enumerate) Subsystems() (it iter.Seq[string], err error) {
	e.lock()
	defer e.unlock()
	if C.udev_enumerate_scan_subsystems(e.ptr) < 0 {
		err = errors.New("udev: udev_enumerate_scan_devices failed")
		return
	}

	return enumerateName(&e.udevContext, func() *C.struct_udev_list_entry {
		e.lock()
		defer e.unlock()
		return C.udev_enumerate_get_list_entry(e.ptr)
	}), nil
}
