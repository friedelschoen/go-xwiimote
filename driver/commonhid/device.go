package commonhid

import (
	"errors"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/internal/common"
)

type commonEvent struct {
	iface     wiimote.Interface
	timestamp time.Time
}

func (e commonEvent) Interface() wiimote.Interface { return e.iface }
func (e commonEvent) Timestamp() time.Time         { return e.timestamp }

type ifaceCore struct {
	dev    *Device
	opened bool
}

func (i *ifaceCore) Close() error {
	i.opened = false
	return nil
}
func (i *ifaceCore) Kind() wiimote.InterfaceKind {
	return wiimote.InterfaceCore
}
func (i *ifaceCore) Device() wiimote.Device {
	return i.dev
}
func (i *ifaceCore) Opened() bool {
	return i.opened
}

type Device struct {
	wiimote.Poller[wiimote.Event]

	transport Transport

	openIfs  map[wiimote.InterfaceKind]wiimote.Interface
	availIfs wiimote.InterfaceKind

	// last known button state (16-bit)
	btnPrev uint16

	// queued events (because a single report can generate multiple key events)
	moreEvents chan wiimote.Event
}

type Transport interface {
	io.ReadWriteCloser
	FD() int
}

func NewDevice(transport Transport) wiimote.Device {
	d := &Device{
		transport:  transport,
		openIfs:    make(map[wiimote.InterfaceKind]wiimote.Interface),
		availIfs:   wiimote.InterfaceCore, // minimal promise for now
		moreEvents: make(chan wiimote.Event, 64),
	}
	d.Poller = common.NewPoller(d)
	return d
}

func (d *Device) Syspath() string { return "" }

func (d *Device) String() string {
	return "wiimote-device (commonhid)"
}

func (d *Device) OpenInterfaces(ifaces wiimote.InterfaceKind, wr bool) error {
	_ = wr // transport might be read-only or read-write; we don't enforce here

	// For now: only Core is implemented.
	if ifaces&wiimote.InterfaceCore != 0 {
		if _, ok := d.openIfs[wiimote.InterfaceCore]; !ok {
			core := &ifaceCore{dev: d, opened: true}
			d.openIfs[wiimote.InterfaceCore] = core

			// optional: notify interface appeared
			d.moreEvents <- &wiimote.EventInterface{
				Event: commonEvent{iface: core, timestamp: time.Now()},
				Kind:  wiimote.InterfaceCore,
			}
		}
	}

	// If caller asked for more: return a meaningful error, but keep Core opened.
	var errs []error
	want := ifaces &^ wiimote.InterfaceCore
	if want != 0 {
		errs = append(errs, os.ErrInvalid)
	}
	return errors.Join(errs...)
}

func (d *Device) FD() int {
	return d.transport.FD()
}

func (d *Device) Interface(kind wiimote.InterfaceKind) wiimote.Interface {
	return d.openIfs[kind]
}

func (d *Device) Available(iface wiimote.InterfaceKind) bool {
	return d.availIfs&iface != 0
}

func (d *Device) LED() (wiimote.Led, error) {
	return 0, os.ErrInvalid
}

func (d *Device) SetLED(leds wiimote.Led) error {
	_ = leds
	return os.ErrInvalid
}

func (d *Device) Battery() (uint, error) {
	return 0, os.ErrInvalid
}

func (d *Device) DevType() (string, error) {
	return "unknown", os.ErrInvalid
}

func (d *Device) Extension() (string, error) {
	return "none", os.ErrInvalid
}

func (d *Device) Poll() (wiimote.Event, bool, error) {
	// 1) return queued events first
	select {
	case ev := <-d.moreEvents:
		return ev, true, nil
	default:
	}

	// 2) read a report
	var buf [64]byte
	n, err := d.transport.Read(buf[:])
	if err != nil {
		// Non-blocking transport support (os.File O_NONBLOCK typically returns EAGAIN)
		if errors.Is(err, syscall.EAGAIN) {
			return nil, false, common.ErrPollAgain
		}

		// Device gone
		if errors.Is(err, io.EOF) {
			return &wiimote.EventGone{
				Event: commonEvent{iface: nil, timestamp: time.Now()},
			}, true, nil
		}

		// Any other error: treat like "watch"/state changed; caller should reopen etc.
		return &wiimote.EventWatch{
			Event: commonEvent{iface: nil, timestamp: time.Now()},
		}, true, nil
	}
	if n <= 0 {
		return nil, false, common.ErrPollAgain
	}

	report := buf[:n]
	ts := time.Now()

	// 3) decode minimal core button report
	// Typical Wiimote "Core Buttons" report is 0x30 and contains:
	// [0]=report id, [1]=buttons high, [2]=buttons low
	//
	// NOTE: If your transport uses a different layout, adjust here.
	if len(report) >= 3 {
		rid := report[0]
		if rid == 0x30 || rid == 0x31 || rid == 0x33 || rid == 0x35 || rid == 0x37 {
			btn := uint16(report[1])<<8 | uint16(report[2])
			d.enqueueButtonDeltas(ts, btn)

			// return first produced event immediately if any
			select {
			case ev := <-d.moreEvents:
				return ev, true, nil
			default:
			}
		}
	}

	// 4) nothing produced from this report (yet)
	return nil, false, common.ErrPollAgain
}

func (d *Device) enqueueButtonDeltas(ts time.Time, btn uint16) {
	prev := d.btnPrev
	if btn == prev {
		return
	}
	d.btnPrev = btn

	core, _ := d.openIfs[wiimote.InterfaceCore].(*ifaceCore)

	emit := func(k wiimote.Key, pressed bool) {
		state := wiimote.StateReleased
		if pressed {
			state = wiimote.StatePressed
		}
		d.moreEvents <- &wiimote.EventKey{
			Event: commonEvent{iface: core, timestamp: ts},
			Code:  k,
			State: state,
		}
	}

	changed := btn ^ prev
	switch {
	case changed&0x0001 != 0:
		emit(wiimote.KeyTwo, btn&0x0001 != 0)
	case changed&0x0002 != 0:
		emit(wiimote.KeyOne, btn&0x0002 != 0)
	case changed&0x0004 != 0:
		emit(wiimote.KeyB, btn&0x0004 != 0)
	case changed&0x0008 != 0:
		emit(wiimote.KeyA, btn&0x0008 != 0)
	case changed&0x0010 != 0:
		emit(wiimote.KeyMinus, btn&0x0010 != 0)
	case changed&0x0080 != 0:
		emit(wiimote.KeyHome, btn&0x0080 != 0)
	case changed&0x0100 != 0:
		emit(wiimote.KeyLeft, btn&0x0100 != 0)
	case changed&0x0200 != 0:
		emit(wiimote.KeyRight, btn&0x0200 != 0)
	case changed&0x0400 != 0:
		emit(wiimote.KeyDown, btn&0x0400 != 0)
	case changed&0x0800 != 0:
		emit(wiimote.KeyUp, btn&0x0800 != 0)
	case changed&0x1000 != 0:
		emit(wiimote.KeyPlus, btn&0x1000 != 0)
	}
}
