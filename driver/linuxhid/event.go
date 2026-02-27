package linuxhid

import (
	"time"

	"github.com/friedelschoen/go-wiimote"
)

type commonEvent struct {
	iface     Interface
	timestamp time.Time
}

func (evt commonEvent) Interface() wiimote.Interface {
	return evt.iface
}

func (evt commonEvent) Timestamp() time.Time {
	return evt.timestamp
}

func (dev *Device) readUmon(pollEv uint32) (wiimote.Event, error) {
	_ = pollEv
	hotplug := false
	remove := false
	path := dev.dev.Syspath()

	// try to merge as many hotplug events as possible
	for {
		ndev := dev.umon.ReceiveDevice()
		if ndev == nil {
			break
		}

		// We are interested in three kinds of events:
		// 1) "change" events on the main HID device notify
		//    us of device-detection events.
		// 2) "remove" events on the main HID device notify
		//    us of device-removal.
		// 3) "add"/"remove" events on input events (not
		//    the evdev events with "devnode") notify us
		//    of extension changes. */

		act := ndev.Action()
		npath := ndev.Syspath()
		node := ndev.Devnode()
		var ppath string
		if p := ndev.Parent(); p != nil && p.Subsystem() == "hid" {
			ppath = p.Syspath()
		}
		if act == "change" && path == npath {
			hotplug = true
		} else if act == "remove" && path == npath {
			remove = true
		} else if node == "" && path == ppath {
			hotplug = true
		}
	}

	// notify caller of removals via special event
	if remove {
		dev.readNodes()
		return &wiimote.EventGone{
			Event: commonEvent{
				timestamp: time.Now(),
				iface:     nil,
			},
		}, nil
	}

	// notify caller via generic hotplug event
	if hotplug {
		dev.readNodes()
		return &wiimote.EventWatch{
			Event: commonEvent{
				timestamp: time.Now(),
				iface:     nil,
			},
		}, nil
	}

	return nil, nil
}

func (dev *Device) dispatchEvent(evFd int32, pollEv uint32) (wiimote.Event, error) {
	if dev.umon != nil && dev.umon.FD() == int(evFd) {
		return dev.readUmon(pollEv)
	}
	for _, iff := range dev.ifs {
		if int32(iff.fd()) != evFd {
			continue
		}
		return dispatchEvent(iff)
	}

	return nil, nil
}
