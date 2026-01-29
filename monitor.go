package xwiimote

import (
	"os"
	"syscall"

	"github.com/friedelschoen/go-xwiimote/pkg/udev"
)

// MonitorType describes how a monitor or enumerator should look for devices.
type MonitorType uint

const (
	// Monitor uses kernel uevents
	MonitorKernel MonitorType = 1
	// Monitor uses udevd
	MonitorUdev MonitorType = 0
)

func (t MonitorType) Name() string {
	switch t {
	case MonitorKernel:
		return "kernel"
	default:
		return "udev"
	}
}

// Monitor describes a monitor for xwiimote-devices. This includes currently available
// but also hot-plugged devices.
//
// Monitors are not thread-safe.
type Monitor struct {
	poller[*Device]
	monitor *udev.Monitor
	enum    chan struct {
		dev *Device
		err error
	}
}

// NewMonitor creates a new monitor.
//
// A monitor always provides all devices that are available on a system
// and hot-plugged devices.
//
// The object and underlying structure is freed automatically by default.
func NewMonitor(typ MonitorType) (*Monitor, error) {
	var mon Monitor
	mon.poller = newPoller(&mon)

	devs, err := IterDevices()
	if err != nil {
		return nil, err
	}
	mon.enum = make(chan struct {
		dev *Device
		err error
	})
	go func() {
		for dev, err := range devs {
			mon.enum <- struct {
				dev *Device
				err error
			}{dev, err}
		}
		close(mon.enum)
	}()

	mon.monitor = udev.NewMonitorFromNetlink(typ.Name())
	if mon.monitor == nil {
		return nil, os.ErrInvalid
	}
	if err := mon.monitor.FilterAddMatchSubsystem("hid"); err != nil {
		return nil, err
	}
	if err := mon.monitor.EnableReceiving(); err != nil {
		return nil, err
	}
	return &mon, nil
}

// FD returns the file-descriptor to notify readiness. The FD is non-blocking.
// Only one file-descriptor exists, that is, this function always returns the
// same descriptor.
func (mon *Monitor) FD() int {
	fd := mon.monitor.GetFD()
	syscall.SetNonblock(fd, false)
	return fd
}

// Poll returns a single device-name on each call. A device-name is actually
// an absolute sysfs path to the device's root-node. This is normally a path
// to /sys/bus/hid/devices/[dev]/. You can use this path to create a new
// struct xwii_iface object.
//
// After a monitor was created, this function returns all currently available
// devices. After all devices have been returned. After that, this function polls the
// monitor for hotplug events and returns hotplugged devices,
// if the monitor was opened to watch the system for hotplug events.
//
// Use FD() to get notified when a new event is available.
func (mon *Monitor) Poll() (*Device, bool, error) {
	// test if enumerator has devices, then wait for new devices
	if iter, ok := <-mon.enum; ok {
		return iter.dev, true, iter.err
	}

	dev := mon.monitor.ReceiveDevice()
	if dev == nil {
		return nil, false, ErrPollAgain
	}
	if (dev.Action() != "" && dev.Action() != "add") || dev.Driver() != "wiimote" || dev.Subsystem() != "hid" {
		return nil, false, ErrPollAgain
	}
	iff, err := newDeviceFromUdev(dev)
	return iff, false, err
}
