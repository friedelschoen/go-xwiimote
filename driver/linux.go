//go:build linux

package driver

import (
	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/driver/commonhid"
	"github.com/friedelschoen/go-wiimote/driver/linuxhidraw"
	"github.com/friedelschoen/go-wiimote/driver/linuxkernel"
	"github.com/friedelschoen/go-wiimote/driver/udev"
)

func NewEnumerate() wiimote.DeviceEnumerator {
	return udev.NewEnumerate()
}

func NewMonitor() wiimote.DeviceMonitor {
	return udev.NewMonitorFromNetlink(udev.MonitorUdev)
}

func NewDevice(info wiimote.DeviceInfo, backend Backend) (wiimote.Device, error) {
	switch backend {
	case BackendKernel:
		return linuxkernel.NewDevice(info, NewMonitor, NewEnumerate)
	default:
		transport, err := linuxhidraw.NewTransportFromInfo(info)
		if err != nil {
			return nil, err
		}
		return commonhid.NewDevice(transport), nil
	}
}
