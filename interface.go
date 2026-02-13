package xwiimote

// #include "input-defs.h"
import "C"
import (
	"io"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type Interface interface {
	// Name of this device,
	Name() string

	// Node is an absolute path which points to the sys-directory. When opened, a node is bound to this interface.
	Node() string

	// Device is the parent of this interface. When opened, a device is bound to this interface.
	Device() *Device

	// Opened returns a bitmask of opened interfaces. Interfaces may be closed due to
	// error-conditions at any time. However, interfaces are never opened
	// automatically.
	//
	// You will get notified whenever this bitmask changes, except on explicit
	// calls to Open() and Close(). See the EventWatch event for more information.
	Opened() bool

	fd() *os.File
	open(dev *Device, node string, wr bool) error
	close() error
}

type commonInterface struct {
	// parent commoniface.device
	dev *Device
	//iface.device node as /dev/input/eventX or ""
	node string
	// Open file or nil
	file *os.File
}

// Node is an absolute path which points to the sys-directory. When opened, a node is bound to this device.
func (iface *commonInterface) Node() string {
	return iface.node
}

func (iface *commonInterface) Device() *Device {
	return iface.dev
}
func (iface *commonInterface) fd() *os.File {
	return iface.file
}

// Opened returns a bitmask of opened interfaces. Interfaces may be closed due to
// error-conditions at any time. However, interfaces are never opened
// automatically.
//
// You will get notified whenever this bitmask changes, except on explicit
// calls to Open() and Close(). See the EventWatch event for more information.
func (iface *commonInterface) Opened() bool {
	return iface.file != nil
}

func (iff *commonInterface) open(dev *Device, node string, wr bool) error {
	if iff.dev != nil && iff.node != "" && iff.file != nil {
		return nil
	}

	iff.dev = dev
	iff.node = node

	flags := syscall.O_NONBLOCK | syscall.O_CLOEXEC
	if wr {
		flags |= os.O_RDWR
	}
	fd, err := os.OpenFile(iff.Node(), flags, 0)
	if err != nil {
		return err
	}

	// name, err := devname(fd)
	// if err != nil {
	// 	return err
	// }
	// if name != iff..Name() {
	// 	return fmt.Errorf("device does not hold correct name: expected %q, got %q", iff.self.Name(), name)
	// }

	var ep syscall.EpollEvent
	ep.Events = syscall.EPOLLIN
	ep.Fd = int32(fd.Fd())
	if err := syscall.EpollCtl(iff.dev.efd, syscall.EPOLL_CTL_ADD, int(fd.Fd()), &ep); err != nil {
		fd.Close()
		return err
	}

	iff.file = fd
	return nil
}

func (iff *commonInterface) close() error {
	if iff.file == nil {
		return nil
	}
	if err := syscall.EpollCtl(iff.dev.efd, syscall.EPOLL_CTL_DEL, int(iff.file.Fd()), nil); err != nil {
		return nil
	}
	if err := iff.file.Close(); err != nil {
		return nil
	}
	iff.file = nil
	return nil
}

type rumbleInterface struct {
	commonInterface

	//  rumble-id for base-core interface force-feedback or -1
	rumbleValid bool
	rumbleID    int
}

func (iface *rumbleInterface) open(dev *Device, node string, wr bool) error {
	if err := iface.commonInterface.open(dev, node, wr); err != nil {
		return err
	}

	return iface.uploadRumble()
}

// Upload the generic rumble event to the device. This may later be used for
// force-feedback effects. The event id is safed for later use.
func (iface *rumbleInterface) uploadRumble() error {
	effect := C.struct_ff_effect{
		_type: C.FF_RUMBLE,
		id:    -1,
	}

	rmb := (*C.struct_ff_rumble_effect)(unsafe.Pointer(&effect.u))
	rmb.strong_magnitude = 1

	if err := ioctl(iface.file.Fd(), C.EVIOCSFF, uintptr(unsafe.Pointer(&effect))); err != nil {
		return err
	}
	iface.rumbleValid = true
	iface.rumbleID = int(effect.id)
	return nil
}

func (iff *rumbleInterface) close() error {
	iff.rumbleValid = false

	return iff.commonInterface.close()
}

// Rumble sets the rumble motor.
//
// This requires the core-interface to be opened in writable mode.
func (dev *rumbleInterface) Rumble(state bool) error {
	if dev.file == nil || !dev.rumbleValid {
		return os.ErrInvalid
	}

	var ev C.struct_input_event
	ev._type = C.EV_FF
	ev.code = C.ushort(dev.rumbleID)
	if state {
		ev.value = 1
	}

	buf := unsafe.Slice((*byte)(unsafe.Pointer(&ev)), unsafe.Sizeof(ev))

	n, err := dev.file.Write(buf)
	if err != nil {
		return err
	}
	if n != int(unsafe.Sizeof(ev)) {
		return io.ErrShortWrite
	}
	return nil

}

type InterfaceCore struct {
	rumbleInterface
}

func (InterfaceCore) Name() string {
	return "Nintendo Wii Remote"
}

func (iface *InterfaceCore) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
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
	ev.iface = iface
	ev.timestamp = ts
	ev.Code = key
	ev.State = KeyState(value)
	return &ev, nil
}

type InterfaceAccel struct {
	commonInterface

	accel Vec3
}

func (InterfaceAccel) Name() string {
	return "Nintendo Wii Remote Accelerometer"
}

func (iface *InterfaceAccel) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event == C.EV_SYN {
		var ev EventAccel
		ev.iface = iface
		ev.timestamp = ts
		ev.Accel = iface.accel
		return &ev, nil
	}

	if event != C.EV_ABS {
		return nil, nil
	}

	switch code {
	case C.ABS_RX:
		iface.accel.X = value
	case C.ABS_RY:
		iface.accel.Y = value
	case C.ABS_RZ:
		iface.accel.Z = value
	}
	return nil, nil
}

type InterfaceIR struct {
	commonInterface

	slots [4]IRSlot
}

func (InterfaceIR) Name() string {
	return "Nintendo Wii Remote IR"
}

func (iface *InterfaceIR) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event == C.EV_SYN {
		var ev EventIR
		ev.iface = iface
		ev.timestamp = ts
		ev.Slots = iface.slots
		return &ev, nil
	}

	if event != C.EV_ABS {
		return nil, nil
	}

	switch code {
	case C.ABS_HAT0X:
		iface.slots[0].X = value
	case C.ABS_HAT0Y:
		iface.slots[0].Y = value
	case C.ABS_HAT1X:
		iface.slots[1].X = value
	case C.ABS_HAT1Y:
		iface.slots[1].Y = value
	case C.ABS_HAT2X:
		iface.slots[2].X = value
	case C.ABS_HAT2Y:
		iface.slots[2].Y = value
	case C.ABS_HAT3X:
		iface.slots[3].X = value
	case C.ABS_HAT3Y:
		iface.slots[3].Y = value
	}
	return nil, nil
}

type InterfaceMotionPlus struct {
	commonInterface

	speed Vec3
}

func (InterfaceMotionPlus) Name() string {
	return "Nintendo Wii Remote Motion Plus"
}

func (iface *InterfaceMotionPlus) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event == C.EV_SYN {
		iface.speed.X -= iface.dev.mpNormalizer.X / 100
		iface.speed.Y -= iface.dev.mpNormalizer.Y / 100
		iface.speed.Z -= iface.dev.mpNormalizer.Z / 100
		if iface.speed.X > 0 {
			iface.dev.mpNormalizer.X += iface.dev.mpNormaizeFactor
		} else {
			iface.dev.mpNormalizer.X -= iface.dev.mpNormaizeFactor
		}
		if iface.speed.Y > 0 {
			iface.dev.mpNormalizer.Y += iface.dev.mpNormaizeFactor
		} else {
			iface.dev.mpNormalizer.Y -= iface.dev.mpNormaizeFactor
		}
		if iface.speed.Z > 0 {
			iface.dev.mpNormalizer.Z += iface.dev.mpNormaizeFactor
		} else {
			iface.dev.mpNormalizer.Z -= iface.dev.mpNormaizeFactor
		}

		var ev EventMotionPlus
		ev.iface = iface
		ev.timestamp = ts
		ev.Speed = iface.speed
		return &ev, nil
	}

	if event != C.EV_ABS {
		return nil, nil
	}

	switch code {
	case C.ABS_RX:
		iface.speed.X = value
	case C.ABS_RY:
		iface.speed.Y = value
	case C.ABS_RZ:
		iface.speed.Z = value
	}

	return nil, nil
}

type InterfaceNunchuck struct {
	commonInterface

	stick Vec2
	accel Vec3
}

func (InterfaceNunchuck) Name() string {
	return "Nintendo Wii Remote Nunchuk"
}

func (iface *InterfaceNunchuck) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
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
		ev.iface = iface
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_HAT0X:
			iface.stick.X = value
		case C.ABS_HAT0Y:
			iface.stick.Y = value
		case C.ABS_RX:
			iface.accel.X = value
		case C.ABS_RY:
			iface.accel.Y = value
		case C.ABS_RZ:
			iface.accel.Z = value
		}
	case C.EV_SYN:
		var ev EventNunchukMove
		ev.iface = iface
		ev.timestamp = ts
		ev.Stick = iface.stick
		ev.Accel = iface.accel
		return &ev, nil
	}

	return nil, nil
}

type InterfaceClassicController struct {
	commonInterface

	stickLeft     Vec2
	stickRight    Vec2
	shoulderLeft  int32
	shoulderRight int32
}

func (InterfaceClassicController) Name() string {
	return "Nintendo Wii Remote Classic Controller"
}

func (iface *InterfaceClassicController) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
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
		ev.iface = iface
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_HAT1X:
			iface.stickLeft.X = value
		case C.ABS_HAT1Y:
			iface.stickLeft.Y = value
		case C.ABS_HAT2X:
			iface.stickRight.X = value
		case C.ABS_HAT2Y:
			iface.stickRight.Y = value
		case C.ABS_HAT3X:
			iface.shoulderLeft = value
		case C.ABS_HAT3Y:
			iface.shoulderRight = value
		}
	case C.EV_SYN:
		var ev EventClassicControllerMove
		ev.iface = iface
		ev.timestamp = ts
		ev.StickLeft = iface.stickLeft
		ev.StickRight = iface.stickRight
		ev.ShoulderLeft = iface.shoulderLeft
		ev.ShoulderRight = iface.shoulderRight
		return &ev, nil
	}

	return nil, nil
}

type balanceboardInterface struct {
	commonInterface

	weights [4]int32
}

func (balanceboardInterface) Name() string {
	return "Nintendo Wii Remote Balance Board"
}

func (iface *balanceboardInterface) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
	if event == C.EV_SYN {
		var ev EventBalanceBoard
		ev.iface = iface
		ev.timestamp = ts
		ev.Weights = iface.weights
		return &ev, nil
	}

	if event != C.EV_ABS {
		return nil, nil
	}

	switch code {
	case C.ABS_HAT0X:
		iface.weights[0] = value
	case C.ABS_HAT0Y:
		iface.weights[1] = value
	case C.ABS_HAT1X:
		iface.weights[2] = value
	case C.ABS_HAT1Y:
		iface.weights[3] = value
	}

	return nil, nil
}

type InterfaceProController struct {
	rumbleInterface

	sticks [2]Vec2
}

func (InterfaceProController) Name() string {
	return "Nintendo Wii Remote Pro Controller"
}

func (iface *InterfaceProController) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
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
		ev.iface = iface
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_X:
			iface.sticks[0].X = value
		case C.ABS_Y:
			iface.sticks[0].Y = value
		case C.ABS_RX:
			iface.sticks[1].X = value
		case C.ABS_RY:
			iface.sticks[1].Y = value
		}
	case C.EV_SYN:
		var ev EventProControllerMove
		ev.iface = iface
		ev.timestamp = ts
		ev.Sticks = iface.sticks
		return &ev, nil
	}

	return nil, nil
}

type InterfaceDrums struct {
	commonInterface

	pad         Vec2
	cymbalLeft  int32
	cymbalRight int32
	tomLeft     int32
	tomRight    int32
	tomFarRight int32
	bass        int32
	hiHat       int32
}

func (InterfaceDrums) Name() string {
	return "Nintendo Wii Remote Drums"
}

func (iface *InterfaceDrums) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
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
		ev.iface = iface
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_X:
			iface.pad.X = value
		case C.ABS_Y:
			iface.pad.Y = value
		case C.ABS_CYMBAL_LEFT:
			iface.cymbalLeft = value
		case C.ABS_CYMBAL_RIGHT:
			iface.cymbalRight = value
		case C.ABS_TOM_LEFT:
			iface.tomLeft = value
		case C.ABS_TOM_RIGHT:
			iface.tomRight = value
		case C.ABS_TOM_FAR_RIGHT:
			iface.tomFarRight = value
		case C.ABS_BASS:
			iface.bass = value
		case C.ABS_HI_HAT:
			iface.hiHat = value
		}
	case C.EV_SYN:
		var ev EventDrumsMove
		ev.iface = iface
		ev.timestamp = ts
		ev.Pad = iface.pad
		ev.CymbalLeft = iface.cymbalLeft
		ev.CymbalRight = iface.cymbalRight
		ev.TomLeft = iface.tomLeft
		ev.TomRight = iface.tomRight
		ev.TomFarRight = iface.tomFarRight
		ev.Bass = iface.bass
		ev.HiHat = iface.hiHat
		return &ev, nil
	}

	return nil, nil
}

type InterfaceGuitar struct {
	commonInterface

	stick     Vec2
	whammyBar int32
	fretBar   int32
}

func (InterfaceGuitar) Name() string {
	return "Nintendo Wii Remote Guitar"
}

func (iface *InterfaceGuitar) acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error) {
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
		ev.iface = iface
		ev.timestamp = ts
		ev.Code = key
		ev.State = KeyState(value)
		return &ev, nil
	case C.EV_ABS:
		switch code {
		case C.ABS_X:
			iface.stick.X = value
		case C.ABS_Y:
			iface.stick.Y = value
		case C.ABS_WHAMMY_BAR:
			iface.whammyBar = value
		case C.ABS_FRET_BOARD:
			iface.fretBar = value
		}
	case C.EV_SYN:
		var ev EventGuitarMove
		ev.iface = iface
		ev.timestamp = ts
		ev.Stick = iface.stick
		ev.WhammyBar = iface.whammyBar
		ev.FretBar = iface.fretBar
		return &ev, nil
	}

	return nil, nil
}

func GetInterface(name string) Interface {
	switch name {
	case "Nintendo Wii Remote":
		return &InterfaceCore{}
	case "Nintendo Wii Remote Accelerometer":
		return &InterfaceAccel{}
	case "Nintendo Wii Remote IR":
		return &InterfaceIR{}
	case "Nintendo Wii Remote Motion Plus":
		return &InterfaceMotionPlus{}
	case "Nintendo Wii Remote Nunchuk":
		return &InterfaceNunchuck{}
	case "Nintendo Wii Remote Classic Controller":
		return &InterfaceClassicController{}
	case "Nintendo Wii Remote Balance Board":
		return &balanceboardInterface{}
	case "Nintendo Wii Remote Pro Controller":
		return &InterfaceProController{}
	case "Nintendo Wii Remote Drums":
		return &InterfaceDrums{}
	case "Nintendo Wii Remote Guitar":
		return &InterfaceGuitar{}
	default:
		return nil
	}
}

type eventInterface interface {
	Interface
	acceptEvent(ts time.Time, event, code uint16, value int32) (Event, error)
}

func readEvent(fd io.Reader) (*C.struct_input_event, error) {
	var ev C.struct_input_event
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&ev)), unsafe.Sizeof(ev))

	n, err := io.ReadFull(fd, buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	if n != int(unsafe.Sizeof(ev)) {
		return nil, io.ErrShortBuffer
	}
	return &ev, nil
}

func dispatchEvent(iff eventInterface) (Event, error) {
	for {
		input, err := readEvent(iff.fd())
		if err != nil {
			iff.Device().CloseInterfaces(iff)
			return &EventWatch{commonEvent{iff, time.Now()}}, nil
		}
		if input == nil {
			return nil, ErrPollAgain
		}
		ts := cTime(input.time)
		eventType := uint16(input._type)
		code := uint16(input.code)
		value := int32(input.value)

		event, err := iff.acceptEvent(ts, eventType, code, value)
		if event != nil || err != nil {
			return event, err
		}
	}
}
