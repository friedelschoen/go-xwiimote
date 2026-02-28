package wiimote

import "iter"

type DeviceInfo interface {
	// Parent returns the parent Device, or nil if the receiver has no parent Device
	Parent() DeviceInfo

	// Subsystem returns the subsystem string of the udev device.
	// The string does not contain any "/".
	Subsystem() string

	// Sysname returns the sysname of the udev device (e.g. ttyS3, sda1...).
	Sysname() string

	// Syspath returns the sys path of the udev device.
	// The path is an absolute path and starts with the sys mount point.
	Syspath() string

	// Devnode returns the device node file name belonging to the udev device.
	// The path is an absolute path, and starts with the device directory.
	Devnode() string

	// Driver returns the driver for the receiver
	Driver() string

	// Action returns the action for the event.
	// This is only valid if the device was received through a monitor.
	// Devices read from sys do not have an action string.
	// Usual actions are: add, remove, change, online, offline.
	Action() string

	// SysattrValue retrieves the content of a sys attribute file, and returns an empty string if there is no sys attribute value.
	// The retrieved value is cached in the device.
	// Repeated calls will return the same value and not open the attribute again.
	SysattrValue(sysattr string) string
}

type DeviceEnumerator interface {
	// AddMatchSubsystem adds a filter for a subsystem of the device to include in the list.
	AddMatchSubsystem(subsystem string) (err error)

	// AddNomatchSubsystem adds a filter for a subsystem of the device to exclude from the list.
	AddNomatchSubsystem(subsystem string) (err error)

	// AddMatchSysattr adds a filter for a sys attribute at the device to include in the list.
	AddMatchSysattr(sysattr, value string) (err error)

	// AddNomatchSysattr adds a filter for a sys attribute at the device to exclude from the list.
	AddNomatchSysattr(sysattr, value string) (err error)

	// AddMatchSysname adds a filter for the name of the device to include in the list.
	AddMatchSysname(sysname string) (err error)

	// AddMatchParent adds a filter for a parent Device to include in the list.
	AddMatchParent(parent DeviceInfo) error

	// AddSyspath adds a device to the list of enumerated devices, to retrieve it back sorted in dependency order.
	AddSyspath(syspath string) (err error)

	// Devices returns an Iterator over the device syspaths matching the filter, sorted in dependency order.
	// The Iterator is using the github.com/jkeiser/iter package.
	Devices() (it iter.Seq[DeviceInfo], err error)

	// Subsystems returns an Iterator over the subsystem syspaths matching the filter, sorted in dependency order.
	// The Iterator is using the github.com/jkeiser/iter package.
	Subsystems() (it iter.Seq[string], err error)
}

type DeviceMonitor interface {
	// FD receives a file descriptor which can be checked for rediness
	FD() int

	EnableReceiving() (err error)

	ReceiveDevice() DeviceInfo

	// SetReceiveBufferSize sets the size of the kernel socket buffer.
	// This call needs the appropriate privileges to succeed.
	SetReceiveBufferSize(size int) (err error)

	// FilterAddMatchSubsystem adds a filter matching the device against a subsystem.
	// This filter is efficiently executed inside the kernel, and libudev subscribers will usually not be woken up for devices which do not match.
	// The filter must be installed before the monitor is switched to listening mode with the DeviceChan function.
	FilterAddMatchSubsystem(subsystem string) (err error)

	// FilterUpdate updates the installed socket filter.
	// This is only needed, if the filter was removed or changed.
	FilterUpdate() (err error)

	// FilterRemove removes all filter from the Monitor.
	FilterRemove() (err error)
}
