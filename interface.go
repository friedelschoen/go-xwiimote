package xwiimote

// #include "input-defs.h"
import "C"
import (
	"fmt"
	"os"
	"syscall"
	"time"
)

type Interface interface {
	Node() string
	FD() *os.File
	open(wr bool) error
	acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error)
}

type commonInterface struct {
	// parent commoniface.device
	dev *Device
	// type of this interface
	typ InterfaceType
	//iface.device node as /dev/input/eventX or ""
	node string
	// Open file or nil
	fd *os.File
}

func (iface *commonInterface) Node() string {
	return iface.node
}

func (iface *commonInterface) FD() *os.File {
	return iface.fd
}

func (iff *commonInterface) open(wr bool) error {
	if iff.fd != nil {
		return nil
	}

	flags := syscall.O_NONBLOCK | syscall.O_CLOEXEC
	if wr {
		flags |= os.O_RDWR
	}
	fd, err := os.OpenFile(iff.Node(), flags, 0)
	if err != nil {
		return err
	}

	name, err := devname(fd)
	if err != nil {
		return err
	}
	if name != iff.typ.Name() {
		return fmt.Errorf("device does not hold correct name: expected %q, got %q", iff.typ.Name(), name)
	}

	var ep syscall.EpollEvent
	ep.Events = syscall.EPOLLIN
	ep.Fd = int32(fd.Fd())
	if err := syscall.EpollCtl(iff.dev.efd, syscall.EPOLL_CTL_ADD, int(fd.Fd()), &ep); err != nil {
		fd.Close()
		return err
	}

	iff.fd = fd
	return nil
}

type coreInterface struct {
	commonInterface
}

func (iface *coreInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event != C.EV_KEY {
		return nil, nil
	}

	if value < 0 || value > 2 {
		return nil, nil
	}

	var key Key
	switch code {
	case C.KEY_LEFT:
		key = KeyLeft
	case C.KEY_RIGHT:
		key = KeyRight
	case C.KEY_UP:
		key = KeyUp
	case C.KEY_DOWN:
		key = KeyDown
	case C.KEY_NEXT:
		key = KeyPlus
	case C.KEY_PREVIOUS:
		key = KeyMinus
	case C.BTN_1:
		key = KeyOne
	case C.BTN_2:
		key = KeyTwo
	case C.BTN_A:
		key = KeyA
	case C.BTN_B:
		key = KeyB
	case C.BTN_MODE:
		key = KeyHome
	default:
		return nil, nil
	}

	var ev EventKey
	ev.timestamp = ts
	ev.Code = key
	ev.State = KeyState(value)
	return &ev, nil
}

type accelInterface struct {
	commonInterface
}

func (iface *accelInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event == C.EV_SYN {
		iface.dev.accelCache.timestamp = ts
		return &iface.dev.accelCache, nil
	}

	if event != C.EV_ABS {
		return nil, nil
	}

	switch code {
	case C.ABS_RX:
		iface.dev.accelCache.Accel.X = value
	case C.ABS_RY:
		iface.dev.accelCache.Accel.Y = value
	case C.ABS_RZ:
		iface.dev.accelCache.Accel.Z = value
	}
	return nil, nil
}

type irInterface struct {
	commonInterface
}

func (iface *irInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event == C.EV_SYN {
		iface.dev.irCache.timestamp = ts
		return &iface.dev.irCache, nil
	}

	if event != C.EV_ABS {
		return nil, nil
	}

	switch code {
	case C.ABS_HAT0X:
		iface.dev.irCache.Slots[0].X = value
	case C.ABS_HAT0Y:
		iface.dev.irCache.Slots[0].Y = value
	case C.ABS_HAT1X:
		iface.dev.irCache.Slots[1].X = value
	case C.ABS_HAT1Y:
		iface.dev.irCache.Slots[1].Y = value
	case C.ABS_HAT2X:
		iface.dev.irCache.Slots[2].X = value
	case C.ABS_HAT2Y:
		iface.dev.irCache.Slots[2].Y = value
	case C.ABS_HAT3X:
		iface.dev.irCache.Slots[3].X = value
	case C.ABS_HAT3Y:
		iface.dev.irCache.Slots[3].Y = value
	}
	return nil, nil
}

type motionplusInterface struct {
	commonInterface
}

func (iface *motionplusInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event == C.EV_SYN {
		iface.dev.mpCache.timestamp = ts

		iface.dev.mpCache.Speed.X -= iface.dev.mpNormalizer.X / 100
		iface.dev.mpCache.Speed.Y -= iface.dev.mpNormalizer.Y / 100
		iface.dev.mpCache.Speed.Z -= iface.dev.mpNormalizer.Z / 100
		if iface.dev.mpCache.Speed.X > 0 {
			iface.dev.mpNormalizer.X += iface.dev.mpNormaizeFactor
		} else {
			iface.dev.mpNormalizer.X -= iface.dev.mpNormaizeFactor
		}
		if iface.dev.mpCache.Speed.Y > 0 {
			iface.dev.mpNormalizer.Y += iface.dev.mpNormaizeFactor
		} else {
			iface.dev.mpNormalizer.Y -= iface.dev.mpNormaizeFactor
		}
		if iface.dev.mpCache.Speed.Z > 0 {
			iface.dev.mpNormalizer.Z += iface.dev.mpNormaizeFactor
		} else {
			iface.dev.mpNormalizer.Z -= iface.dev.mpNormaizeFactor
		}

		return &iface.dev.mpCache, nil
	}

	if event != C.EV_ABS {
		return nil, nil
	}

	switch code {
	case C.ABS_RX:
		iface.dev.mpCache.Speed.X = value
	case C.ABS_RY:
		iface.dev.mpCache.Speed.Y = value
	case C.ABS_RZ:
		iface.dev.mpCache.Speed.Z = value
	}

	return nil, nil
}

type nunchukInterface struct {
	commonInterface
}

func (iface *nunchukInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	switch event {
	case C.EV_KEY:
		if value < 0 || value > 2 {
			return nil, nil
		}
		var key Key
		switch code {
		case C.BTN_C:
			key = KeyC
		case C.BTN_Z:
			key = KeyZ
		default:
			return nil, nil
		}

		var ev EventNunchukKey
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_HAT0X:
			iface.dev.nunchukCache.Stick.X = value
		case C.ABS_HAT0Y:
			iface.dev.nunchukCache.Stick.Y = value
		case C.ABS_RX:
			iface.dev.nunchukCache.Accel.X = value
		case C.ABS_RY:
			iface.dev.nunchukCache.Accel.Y = value
		case C.ABS_RZ:
			iface.dev.nunchukCache.Accel.Z = value
		}
	case C.EV_SYN:
		iface.dev.nunchukCache.timestamp = ts
		return &iface.dev.nunchukCache, nil
	}

	return nil, nil
}

type classiccontrollerInterface struct {
	commonInterface
}

func (iface *classiccontrollerInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	switch event {
	case C.EV_KEY:
		if value < 0 || value > 2 {
			return nil, nil
		}

		var key Key
		switch code {
		case C.BTN_A:
			key = KeyA
		case C.BTN_B:
			key = KeyB
		case C.BTN_X:
			key = KeyX
		case C.BTN_Y:
			key = KeyY
		case C.KEY_NEXT:
			key = KeyPlus
		case C.KEY_PREVIOUS:
			key = KeyMinus
		case C.BTN_MODE:
			key = KeyHome
		case C.KEY_LEFT:
			key = KeyLeft
		case C.KEY_RIGHT:
			key = KeyRight
		case C.KEY_UP:
			key = KeyUp
		case C.KEY_DOWN:
			key = KeyDown
		case C.BTN_TL:
			key = KeyTL
		case C.BTN_TR:
			key = KeyTR
		case C.BTN_TL2:
			key = KeyZL
		case C.BTN_TR2:
			key = KeyZR
		default:
			return nil, nil
		}

		var ev EventClassicControllerKey
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_HAT1X:
			iface.dev.classicCache.StickLeft.X = value
		case C.ABS_HAT1Y:
			iface.dev.classicCache.StickLeft.Y = value
		case C.ABS_HAT2X:
			iface.dev.classicCache.StickRight.X = value
		case C.ABS_HAT2Y:
			iface.dev.classicCache.StickRight.Y = value
		case C.ABS_HAT3X:
			iface.dev.classicCache.ShoulderLeft = value
		case C.ABS_HAT3Y:
			iface.dev.classicCache.ShoulderRight = value
		}
	case C.EV_SYN:
		iface.dev.classicCache.timestamp = ts
		return &iface.dev.classicCache, nil
	}

	return nil, nil
}

type balanceboardInterface struct {
	commonInterface
}

func (iface *balanceboardInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event == C.EV_SYN {
		iface.dev.bboardCache.timestamp = ts
		return &iface.dev.bboardCache, nil
	}

	if event != C.EV_ABS {
		return nil, nil
	}

	switch code {
	case C.ABS_HAT0X:
		iface.dev.bboardCache.Weights[0] = value
	case C.ABS_HAT0Y:
		iface.dev.bboardCache.Weights[1] = value
	case C.ABS_HAT1X:
		iface.dev.bboardCache.Weights[2] = value
	case C.ABS_HAT1Y:
		iface.dev.bboardCache.Weights[3] = value
	}

	return nil, nil
}

type procontrollerInterface struct {
	commonInterface
}

func (iface *procontrollerInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	switch event {
	case C.EV_KEY:
		if value < 0 || value > 2 {
			return nil, nil
		}

		var key Key
		switch code {
		case C.BTN_EAST:
			key = KeyA
		case C.BTN_SOUTH:
			key = KeyB
		case C.BTN_NORTH:
			key = KeyX
		case C.BTN_WEST:
			key = KeyY
		case C.BTN_START:
			key = KeyPlus
		case C.BTN_SELECT:
			key = KeyMinus
		case C.BTN_MODE:
			key = KeyHome
		case C.BTN_DPAD_LEFT:
			key = KeyLeft
		case C.BTN_DPAD_RIGHT:
			key = KeyRight
		case C.BTN_DPAD_UP:
			key = KeyUp
		case C.BTN_DPAD_DOWN:
			key = KeyDown
		case C.BTN_TL:
			key = KeyTL
		case C.BTN_TR:
			key = KeyTR
		case C.BTN_TL2:
			key = KeyZL
		case C.BTN_TR2:
			key = KeyZR
		case C.BTN_THUMBL:
			key = KeyThumbL
		case C.BTN_THUMBR:
			key = KeyThumbR
		default:
			return nil, nil
		}

		var ev EventProControllerKey
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_X:
			iface.dev.proCache.Sticks[0].X = value
		case C.ABS_Y:
			iface.dev.proCache.Sticks[0].Y = value
		case C.ABS_RX:
			iface.dev.proCache.Sticks[1].X = value
		case C.ABS_RY:
			iface.dev.proCache.Sticks[1].Y = value
		}
	case C.EV_SYN:
		iface.dev.proCache.timestamp = ts
		return &iface.dev.proCache, nil
	}

	return nil, nil
}

type drumsInterface struct {
	commonInterface
}

func (iface *drumsInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	switch event {
	case C.EV_KEY:
		if value < 0 || value > 2 {
			return nil, nil
		}

		var key Key
		switch code {
		case C.BTN_START:
			key = KeyPlus
		case C.BTN_SELECT:
			key = KeyMinus
		default:
			return nil, nil
		}

		var ev EventDrumsKey
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_X:
			iface.dev.drumsCache.Pad.X = value
		case C.ABS_Y:
			iface.dev.drumsCache.Pad.Y = value
		case C.ABS_CYMBAL_LEFT:
			iface.dev.drumsCache.CymbalLeft = value
		case C.ABS_CYMBAL_RIGHT:
			iface.dev.drumsCache.CymbalRight = value
		case C.ABS_TOM_LEFT:
			iface.dev.drumsCache.TomLeft = value
		case C.ABS_TOM_RIGHT:
			iface.dev.drumsCache.TomRight = value
		case C.ABS_TOM_FAR_RIGHT:
			iface.dev.drumsCache.TomFarRight = value
		case C.ABS_BASS:
			iface.dev.drumsCache.Bass = value
		case C.ABS_HI_HAT:
			iface.dev.drumsCache.HiHat = value
		}
	case C.EV_SYN:
		iface.dev.drumsCache.timestamp = ts
		return &iface.dev.drumsCache, nil
	}

	return nil, nil
}

type guitarInterface struct {
	commonInterface
}

func (iface *guitarInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	switch event {
	case C.EV_KEY:
		if value < 0 || value > 2 {
			return nil, nil
		}

		var key Key
		switch code {
		case C.BTN_FRET_FAR_UP:
			key = KeyFretFarUp
		case C.BTN_FRET_UP:
			key = KeyFretUp
		case C.BTN_FRET_MID:
			key = KeyFretMid
		case C.BTN_FRET_LOW:
			key = KeyFretLow
		case C.BTN_FRET_FAR_LOW:
			key = KeyFretFarLow
		case C.BTN_STRUM_BAR_UP:
			key = KeyStrumBarUp
		case C.BTN_STRUM_BAR_DOWN:
			key = KeyStrumBarDown
		case C.BTN_START:
			key = KeyPlus
		case C.BTN_MODE:
			key = KeyHome
		default:
			return nil, nil
		}

		var ev EventGuitarKey
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_X:
			iface.dev.guitarCache.Stick.X = value
		case C.ABS_Y:
			iface.dev.guitarCache.Stick.Y = value
		case C.ABS_WHAMMY_BAR:
			iface.dev.guitarCache.WhammyBar = value
		case C.ABS_FRET_BOARD:
			iface.dev.guitarCache.FretBar = value
		}
	case C.EV_SYN:
		iface.dev.guitarCache.timestamp = ts
		return &iface.dev.guitarCache, nil
	}

	return nil, nil
}

func (dev *Device) newInterface(iface InterfaceType, node string) Interface {
	var cif commonInterface
	cif.dev = dev
	cif.typ = iface
	cif.node = node

	switch iface {
	case InterfaceCore:
		return &coreInterface{cif}
	case InterfaceAccel:
		return &accelInterface{cif}
	case InterfaceIR:
		return &irInterface{cif}
	case InterfaceMotionPlus:
		return &motionplusInterface{cif}
	case InterfaceNunchuk:
		return &nunchukInterface{cif}
	case InterfaceClassicController:
		return &classiccontrollerInterface{cif}
	case InterfaceBalanceBoard:
		return &balanceboardInterface{cif}
	case InterfaceProController:
		return &procontrollerInterface{cif}
	case InterfaceDrums:
		return &drumsInterface{cif}
	case InterfaceGuitar:
		return &guitarInterface{cif}
	}
	return nil
}
