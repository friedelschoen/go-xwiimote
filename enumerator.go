package xwiimote

// #cgo pkg-config: libxwiimote
import "C"
import (
	"iter"

	"github.com/friedelschoen/go-xwiimote/pkg/udev"
)

// Enumerator describes a single one-time enumerator for xwiimote-devices.
// Enumerators are not thread-safe.
type Enumerator struct {
	udev      udev.Udev
	enumerate *udev.Enumerate

	next func() (string, bool)
	stop func()
}

// NewEnumerator creates a new enumerator.
//
// A monitor always provides all devices that are available on a system.
//
// The object and underlying structure is freed automatically by default.
func NewEnumerator(typ MonitorType) (*Enumerator, error) {
	var enum Enumerator

	enum.enumerate = enum.udev.NewEnumerate()
	if err := enum.enumerate.AddMatchSubsystem("hid"); err != nil {
		return nil, err
	}

	devs, err := enum.enumerate.Devices()
	if err != nil {
		return nil, err
	}
	enum.next, enum.stop = iter.Pull(devs)
	return &enum, nil
}

// Next returns a single device-name on each call. A device-name is actually
// an absolute sysfs path to the device's root-node. This is normally a path
// to /sys/bus/hid/devices/[dev]/. You can use this path to create a new
// Device object. If the enumerator is exhausted an empty string is returned and
// no new elements will be provided.
func (enum *Enumerator) Next() string {
	path, ok := enum.next()
	if !ok {
		return ""
	}
	dev := enum.udev.NewDeviceFromSyspath(path)
	if dev == nil {
		return ""
	}
	if dev.Action() != "add" || dev.Driver() != "wiimote" || dev.Subsystem() != "hid" {
		return ""
	}
	return dev.Syspath()
}

// IterDevices returns all currently available devices. It is a wrapper of Enumerator and is reentrant.
func IterDevices(typ MonitorType) iter.Seq[string] {
	return func(yield func(string) bool) {
		enum, err := NewEnumerator(typ)
		if err != nil {
			return
		}
		for {
			path := enum.Next()
			if path == "" {
				return
			}
			if !yield(path) {
				return
			}
		}
	}
}
