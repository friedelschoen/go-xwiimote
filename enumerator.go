package xwiimote

import (
	"iter"

	"github.com/friedelschoen/go-xwiimote/pkg/udev"
	"github.com/friedelschoen/go-xwiimote/pkg/udev/sequences"
)

// IterDevices returns all currently available devices. It returns an error if the
// initialization failed. Each iteration yields a device and error if the device-creation failed.
func IterDevices() (iter.Seq2[*Device, error], error) {
	enum := udev.NewEnumerate()
	if err := enum.AddMatchSubsystem("hid"); err != nil {
		return nil, err
	}

	iter, err := enum.Devices()
	if err != nil {
		return nil, err
	}

	deviter := sequences.Map(iter, func(path string) *udev.Device {
		dev := udev.NewDeviceFromSyspath(path)
		if dev == nil {
			return nil
		}
		if dev.Action() != "" && dev.Action() != "add" {
			return nil
		}
		if dev.Driver() != "wiimote" || dev.Subsystem() != "hid" {
			return nil
		}
		return dev
	})
	deviter = sequences.Filter(deviter, func(d *udev.Device) bool {
		return d != nil
	})
	return sequences.Map12(deviter, newDeviceFromUdev), nil
}
