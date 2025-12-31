package xwiimote

import (
	"iter"

	"github.com/friedelschoen/go-xwiimote/pkg/udev"
)

// IterDevices returns all currently available devices. It returns an error if the
// initialization failed. Each iteration yields a device and error if the device-creation failed.
func IterDevices() (iter.Seq2[*Device, error], error) {
	var udev udev.Udev
	return iterDevicesWithUdev(&udev)
}

func iterDevicesWithUdev(udev *udev.Udev) (iter.Seq2[*Device, error], error) {
	enum := udev.NewEnumerate()
	if err := enum.AddMatchSubsystem("hid"); err != nil {
		return nil, err
	}

	iter, err := enum.Devices()
	if err != nil {
		return nil, err
	}
	return func(yield func(*Device, error) bool) {
		iter(func(path string) bool {
			dev := udev.NewDeviceFromSyspath(path)
			if dev == nil {
				return true
			}
			if dev.Action() != "add" || dev.Driver() != "wiimote" || dev.Subsystem() != "hid" {
				return true
			}
			return yield(newDeviceWithUdev(udev, dev))
		})
	}, nil
}
