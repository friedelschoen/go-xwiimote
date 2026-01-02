package udev

// #cgo pkg-config: libudev
// #include <libudev.h>
import "C"
import (
	"errors"
)

// Monitor is an opaque object handling an event source
type Monitor struct {
	udevContext
	ptr *C.struct_udev_monitor
}

// Unref the monitor
func monitorUnref(m *Monitor) {
	C.udev_monitor_unref(m.ptr)
	C.udev_unref(m.udevPtr)
}

// GetFD receives a file descriptor which can be checked for rediness
func (m *Monitor) GetFD() int {
	m.lock()
	defer m.unlock()
	fd := C.udev_monitor_get_fd(m.ptr)
	return int(fd)
}

func (m *Monitor) EnableReceiving() (err error) {
	m.lock()
	defer m.unlock()
	if C.udev_monitor_enable_receiving(m.ptr) != 0 {
		err = errors.New("udev: udev_monitor_enable_receiving failed")
	}
	return
}

func (m *Monitor) ReceiveDevice() *Device {
	m.lock()
	defer m.unlock()
	ptr := C.udev_monitor_receive_device(m.ptr)
	if ptr == nil {
		return nil
	}
	d := newDevice()
	d.ptr = ptr
	return d
}

// SetReceiveBufferSize sets the size of the kernel socket buffer.
// This call needs the appropriate privileges to succeed.
func (m *Monitor) SetReceiveBufferSize(size int) (err error) {
	m.lock()
	defer m.unlock()
	if C.udev_monitor_set_receive_buffer_size(m.ptr, (C.int)(size)) != 0 {
		err = errors.New("udev: udev_monitor_set_receive_buffer_size failed")
	}
	return
}

// FilterAddMatchSubsystem adds a filter matching the device against a subsystem.
// This filter is efficiently executed inside the kernel, and libudev subscribers will usually not be woken up for devices which do not match.
// The filter must be installed before the monitor is switched to listening mode with the DeviceChan function.
func (m *Monitor) FilterAddMatchSubsystem(subsystem string) (err error) {
	m.lock()
	defer m.unlock()
	s := C.CString(subsystem)
	defer freeCharPtr(s)
	if C.udev_monitor_filter_add_match_subsystem_devtype(m.ptr, s, nil) != 0 {
		err = errors.New("udev: udev_monitor_filter_add_match_subsystem_devtype failed")
	}
	return
}

// FilterAddMatchSubsystemDevtype adds a filter matching the device against a subsystem and device type.
// This filter is efficiently executed inside the kernel, and libudev subscribers will usually not be woken up for devices which do not match.
// The filter must be installed before the monitor is switched to listening mode with the DeviceChan function.
func (m *Monitor) FilterAddMatchSubsystemDevtype(subsystem, devtype string) (err error) {
	m.lock()
	defer m.unlock()
	s, d := C.CString(subsystem), C.CString(devtype)
	defer freeCharPtr(s)
	defer freeCharPtr(d)
	if C.udev_monitor_filter_add_match_subsystem_devtype(m.ptr, s, d) != 0 {
		err = errors.New("udev: udev_monitor_filter_add_match_subsystem_devtype failed")
	}
	return
}

// FilterAddMatchTag adds a filter matching the device against a tag.
// This filter is efficiently executed inside the kernel, and libudev subscribers will usually not be woken up for devices which do not match.
// The filter must be installed before the monitor is switched to listening mode.
func (m *Monitor) FilterAddMatchTag(tag string) (err error) {
	m.lock()
	defer m.unlock()
	t := C.CString(tag)
	defer freeCharPtr(t)
	if C.udev_monitor_filter_add_match_tag(m.ptr, t) != 0 {
		err = errors.New("udev: udev_monitor_filter_add_match_tag failed")
	}
	return
}

// FilterUpdate updates the installed socket filter.
// This is only needed, if the filter was removed or changed.
func (m *Monitor) FilterUpdate() (err error) {
	m.lock()
	defer m.unlock()
	if C.udev_monitor_filter_update(m.ptr) != 0 {
		err = errors.New("udev: udev_monitor_filter_update failed")
	}
	return
}

// FilterRemove removes all filter from the Monitor.
func (m *Monitor) FilterRemove() (err error) {
	m.lock()
	defer m.unlock()
	if C.udev_monitor_filter_remove(m.ptr) != 0 {
		err = errors.New("udev: udev_monitor_filter_remove failed")
	}
	return
}
