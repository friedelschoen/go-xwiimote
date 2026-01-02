package xwiimote

// #include "input-defs.h"
import "C"
import (
	"fmt"
	"io"
	"os"
	"time"
	"unsafe"
)

// Key Event Identifiers
//
// For each key found on a supported device, a separate key identifier is
// defined. Note that a device may have a specific key (for instance: HOME) on
// the main device and on an extension device. An application can detect which
// key was pressed examining the event-type field.
// Some devices report common keys as both, extension and core events. In this
// case the kernel is required to filter these and you should report it as a
// bug. A single physical key-press should never be reported twice, even on two
// different interfaces.
type Key uint

const (
	KeyLeft Key = iota
	KeyRight
	KeyUp
	KeyDown
	KeyA
	KeyB
	KeyPlus
	KeyMinus
	KeyHome
	KeyOne
	KeyTwo
	KeyX
	KeyY
	KeyTL
	KeyTR
	KeyZL
	KeyZR

	// Left thumb button
	//
	// This is reported if the left analog stick is pressed. Not all analog
	// sticks support this. The Wii-U Pro Controller is one of few devices
	// that report this event.
	KeyThumbL

	// Right thumb button
	//
	// This is reported if the right analog stick is pressed. Not all analog
	// sticks support this. The Wii-U Pro Controller is one of few devices
	// that report this event.
	KeyThumbR

	// Extra C button
	//
	// This button is not part of the standard action pad but reported by
	// extension controllers like the Nunchuk. It is supposed to extend the
	// standard A and B buttons.
	KeyC

	// Extra Z button
	//
	// This button is not part of the standard action pad but reported by
	// extension controllers like the Nunchuk. It is supposed to extend the
	// standard X and Y buttons.
	KeyZ

	// Guitar Strum-bar-up event
	//
	// Emitted by guitars if the strum-bar is moved up.
	KeyStrumBarUp

	// Guitar Strum-bar-down event
	//
	// Emitted by guitars if the strum-bar is moved down.
	KeyStrumBarDown

	// Guitar Fret-Far-Up event
	//
	// Emitted by guitars if the upper-most fret-bar is pressed.
	KeyFretFarUp

	// Guitar Fret-Up event
	//
	// Emitted by guitars if the second-upper fret-bar is pressed.
	KeyFretUp

	// Guitar Fret-Mid event
	//
	// Emitted by guitars if the mid fret-bar is pressed.
	KeyFretMid

	// Guitar Fret-Low event
	//
	// Emitted by guitars if the second-lowest fret-bar is pressed.
	KeyFretLow

	// Guitar Fret-Far-Low event
	//
	// Emitted by guitars if the lower-most fret-bar is pressed.
	KeyFretFarLow
)

type KeyState uint

const (
	// The key is released, alternativly KeyUp
	StateReleased KeyState = iota
	// The key is pressed, alternativly KeyDown
	StatePressed
	// The key is hold down and repeats
	StateRepeated
)

// Vec2 represents a 2D point or vector to X and Y, may be interpreted different depending on the event .
type Vec2 struct{ X, Y int32 }

// Vec3 represents a 3D point or vector to X, Y and Z, may be interpreted different depending on the event.
type Vec3 struct{ X, Y, Z int32 }

// Event interface describes an event fired by Device.Dispatch(),
// consider using a type-switch to retrieve the specific event type and data
type Event interface {
	Timestamp() time.Time
}

type commonEvent struct {
	timestamp time.Time
}

func (evt *commonEvent) Timestamp() time.Time {
	return evt.timestamp
}

// EventKey is fired whenever a key is pressed or released. Valid
// key-events include all the events reported by the core-interface,
// which is normally only LEFT, RIGHT, UP, DOWN, A, B, PLUS, MINUS,
// HOME, ONE, TWO.
type EventKey struct {
	commonEvent
	Code  Key
	State KeyState
}

// EventAccel provides accelerometer data.
// Note that the accelerometer reports acceleration data, not speed
// data!
type EventAccel struct {
	commonEvent
	Accel Vec3
}

// EventIR provides IR-camera events. The camera can track up two four IR
// sources. As long as a single source is tracked, it stays at it's
// pre-allocated slot.
//
// Use IRSlot.Valid() to see whether a specific slot is
// currently valid or whether it currently doesn't track any IR source.
type EventIR struct {
	commonEvent
	Slots [4]IRSlot
}

// IRSlot describes Infra-Red Tracking on a WiiMote
type IRSlot struct {
	Vec2
}

// Valid returns wether this slot holds a valid source. If not it has no track and is considered disabled.
func (slot IRSlot) Valid() bool {
	return slot.X != 1023 || slot.Y != 1023
}

// EventBalanceBoard provides balance-board weight data. Four sensors report weight-data
// for each of the four edges of the board.
type EventBalanceBoard struct {
	commonEvent
	Weights [4]int32
}

// EventMotionPlus provides gyroscope events. These describe rotational speed, not
// acceleration, of the motion-plus extension.
type EventMotionPlus struct {
	commonEvent
	Speed Vec3
}

// EventProControllerKey provides button events of the pro-controller
// and are reported via this interface
// and not via the core-interface (which only reports core-buttons).
// Valid buttons include: LEFT, RIGHT, UP, DOWN, PLUS, MINUS, HOME, X,
// Y, A, B, TR, TL, ZR, ZL, THUMBL, THUMBR.
// Payload type is struct xwii_event_key.
type EventProControllerKey struct {
	commonEvent
	Code  Key
	State KeyState
}

// EventProControllerMove provides movement of analog sticks on the pro-controller and is
// reported via this event.
type EventProControllerMove struct {
	commonEvent
	Sticks [2]Vec2
}

// EventWatch is sent whenever an extension was hotplugged (plugged or
// unplugged), a device-detection finished or some other static data
// changed which cannot be monitored separately.
// An application should check what changed by examining the device is
// testing whether all required interfaces are still available.
// Non-hotplug aware devices may discard this event.
//
// This is only returned if you explicitly watched for hotplug events.
// See Device.Watch().
//
// This event is also returned if an interface is closed because the
// kernel closed our file-descriptor (for whatever reason). This is
// returned regardless whether you watch for hotplug events or not.
type EventWatch struct {
	commonEvent
}

// EventClassicControllerKey provides Classic Controller key events.
// Button events of the classic controller are reported via this
// interface and not via the core-interface (which only reports
// core-buttons).
// Valid buttons include: LEFT, RIGHT, UP, DOWN, PLUS, MINUS, HOME, X,
// Y, A, B, TR, TL, ZR, ZL.
type EventClassicControllerKey struct {
	commonEvent
	Code  Key
	State KeyState
}

// EventClassicControllerMove provides Classic Controller movement events.
// Movement of analog sticks are reported via this event. The payload
// is a struct xwii_event_abs and the first two array elements contain
// the absolute x and y position of both analog sticks.
// The x value of the third array element contains the absolute position
// of the TL trigger. The y value contains the absolute position for the
// TR trigger. Note that many classic controllers do not have analog
// TL/TR triggers, in which case these read 0 or MAX (63). The digital
// TL/TR buttons are always reported correctly.
type EventClassicControllerMove struct {
	commonEvent
	StickLeft     Vec2
	StickRight    Vec2
	ShoulderLeft  int32
	ShoulderRight int32
}

// EventNunchukKey provides Nunchuk key events.
// Button events of the nunchuk controller are reported via this
// interface and not via the core-interface (which only reports
// core-buttons).
// Valid buttons include: C, Z
type EventNunchukKey struct {
	commonEvent
	Code  Key
	State KeyState
}

// EventNunchukMove provides Nunchuk movement events.
// Movement events of the nunchuk controller are reported via this
// interface. Payload is of type struct xwii_event_abs. The first array
// element contains the x/y positions of the analog stick. The second
// array element contains the accelerometer information.
type EventNunchukMove struct {
	commonEvent
	Stick Vec2
	Accel Vec3
}

// EventDrumsKey provides Drums key events.
// Button events for drums controllers. Valid buttons are PLUS and MINUS
// for the +/- buttons on the center-bar.
type EventDrumsKey struct {
	commonEvent
	Code  Key
	State KeyState
}

// EventDrumsMove provides Drums movement event
// Movement and pressure events for drums controllers.
type EventDrumsMove struct {
	commonEvent
	Pad         Vec2
	CymbalLeft  int32
	CymbalRight int32
	TomLeft     int32
	TomRight    int32
	TomFarRight int32
	Bass        int32
	HiHat       int32
}

// EventGuitarKey provides Guitar key events
// Button events for guitar controllers. Valid buttons are HOME and PLUS
// for the StarPower/Home button and the + button. Furthermore, you get
// FRET_FAR_UP, FRET_UP, FRET_MID, FRET_LOW, FRET_FAR_LOW for fret
// activity and STRUM_BAR_UP and STRUM_BAR_LOW for the strum bar.
type EventGuitarKey struct {
	commonEvent
	Code  Key
	State KeyState
}

// EventGuitarMove provides Guitar movement events.
// Movement information for guitar controllers.
type EventGuitarMove struct {
	commonEvent
	Stick     Vec2
	WhammyBar int32
	FretBar   int32
}

// EventGone provides Removal Event.
// This event is sent whenever the device was removed. No payload is provided.
// Non-hotplug aware applications may discard this event.
// This is only returned if you explicitly watched for hotplug events.
//
// See Device.Watch().
type EventGone struct {
	commonEvent
}

func (dev *Device) readUmon(pollEv uint32) (Event, error) {
	_ = pollEv
	hotplug := false
	remove := false
	path := dev.dev.Syspath()

	/* try to merge as many hotplug events as possible */
	for {
		ndev := dev.umon.ReceiveDevice()
		if ndev == nil {
			break
		}

		/* We are interested in three kinds of events:
		*  1) "change" events on the main HID device notify
		*     us of device-detection events.
		*  1) "remove" events on the main HID device notify
		*     us of device-removal.
		*  3) "add"/"remove" events on input events (not
		*     the evdev events with "devnode") notify us
		*     of extension changes. */

		act := ndev.Action()
		npath := ndev.Syspath()
		node := ndev.Devnode()
		var ppath string
		if p := ndev.ParentWithSubsystemDevtype("hid", ""); p != nil {
			ppath = p.Syspath()
		}
		fmt.Printf("-- act=%v, npath=%v, node=%v, parent=%v\n", act, npath, node, ppath)
		if act == "change" && path == npath {
			hotplug = true
		} else if act == "remove" && path == npath {
			remove = true
		} else if node == "" && path == ppath {
			hotplug = true
		}
	}

	/* notify caller of removals via special event */
	if remove {
		dev.readNodes()
		return &EventGone{
			commonEvent{
				timestamp: time.Now(),
			},
		}, nil
	}

	/* notify caller via generic hotplug event */
	if hotplug {
		dev.readNodes()
		return &EventWatch{
			commonEvent{
				timestamp: time.Now(),
			},
		}, nil
	}

	return nil, nil
}

func read_event(fd *os.File) (*C.struct_input_event, error) {
	var ev C.struct_input_event
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&ev)), unsafe.Sizeof(ev))

	n, err := fd.Read(buf)
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

func (dev *Device) readCore() (Event, error) {
	fd := dev.ifs[InterfaceCore].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceCore)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type != C.EV_KEY {
		goto try_again
	}

	if input.value < 0 || input.value > 2 {
		goto try_again
	}

	var key Key
	switch input.code {
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
		goto try_again
	}

	var ev EventKey
	ev.timestamp = cTime(input.time)
	ev.Code = key
	ev.State = KeyState(input.value)
	return &ev, nil
}

func (dev *Device) readAccel() (Event, error) {
	fd := dev.ifs[InterfaceAccel].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceAccel)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_SYN {
		dev.accel_cache.timestamp = cTime(input.time)
		return &dev.accel_cache, nil
	}

	if input._type != C.EV_ABS {
		goto try_again
	}

	switch input.code {
	case C.ABS_RX:
		dev.accel_cache.Accel.X = int32(input.value)
	case C.ABS_RY:
		dev.accel_cache.Accel.Y = int32(input.value)
	case C.ABS_RZ:
		dev.accel_cache.Accel.Z = int32(input.value)
	}
	goto try_again
}

func (dev *Device) readIR() (Event, error) {
	fd := dev.ifs[InterfaceIR].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceIR)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_SYN {
		dev.ir_cache.timestamp = cTime(input.time)
		return &dev.ir_cache, nil
	}

	if input._type != C.EV_ABS {
		goto try_again
	}

	switch input.code {
	case C.ABS_HAT0X:
		dev.ir_cache.Slots[0].X = int32(input.value)
	case C.ABS_HAT0Y:
		dev.ir_cache.Slots[0].Y = int32(input.value)
	case C.ABS_HAT1X:
		dev.ir_cache.Slots[1].X = int32(input.value)
	case C.ABS_HAT1Y:
		dev.ir_cache.Slots[1].Y = int32(input.value)
	case C.ABS_HAT2X:
		dev.ir_cache.Slots[2].X = int32(input.value)
	case C.ABS_HAT2Y:
		dev.ir_cache.Slots[2].Y = int32(input.value)
	case C.ABS_HAT3X:
		dev.ir_cache.Slots[3].X = int32(input.value)
	case C.ABS_HAT3Y:
		dev.ir_cache.Slots[3].Y = int32(input.value)
	}
	goto try_again
}

func (dev *Device) readMP() (Event, error) {
	fd := dev.ifs[InterfaceMotionPlus].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceMotionPlus)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_SYN {
		dev.mp_cache.timestamp = cTime(input.time)

		dev.mp_cache.Speed.X -= dev.mp_normalizer.X / 100
		dev.mp_cache.Speed.Y -= dev.mp_normalizer.Y / 100
		dev.mp_cache.Speed.Z -= dev.mp_normalizer.Z / 100
		if dev.mp_cache.Speed.X > 0 {
			dev.mp_normalizer.X += dev.mp_normalize_factor
		} else {
			dev.mp_normalizer.X -= dev.mp_normalize_factor
		}
		if dev.mp_cache.Speed.Y > 0 {
			dev.mp_normalizer.Y += dev.mp_normalize_factor
		} else {
			dev.mp_normalizer.Y -= dev.mp_normalize_factor
		}
		if dev.mp_cache.Speed.Z > 0 {
			dev.mp_normalizer.Z += dev.mp_normalize_factor
		} else {
			dev.mp_normalizer.Z -= dev.mp_normalize_factor
		}

		return &dev.mp_cache, nil
	}

	if input._type != C.EV_ABS {
		goto try_again
	}

	switch input.code {
	case C.ABS_RX:
		dev.mp_cache.Speed.X = int32(input.value)
	case C.ABS_RY:
		dev.mp_cache.Speed.Y = int32(input.value)
	case C.ABS_RZ:
		dev.mp_cache.Speed.Z = int32(input.value)
	}

	goto try_again
}

func (dev *Device) readNunchuck() (Event, error) {
	fd := dev.ifs[InterfaceNunchuk].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceNunchuk)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_KEY {
		if input.value < 0 || input.value > 2 {
			goto try_again
		}
		var key Key
		switch input.code {
		case C.BTN_C:
			key = KeyC
		case C.BTN_Z:
			key = KeyZ
		default:
			goto try_again
		}

		var ev EventNunchukKey
		ev.timestamp = cTime(input.time)
		ev.Code = key
		ev.State = KeyState(input.value)
		return &ev, nil
	} else if input._type == C.EV_ABS {
		switch input.code {
		case C.ABS_HAT0X:
			dev.nunchuk_cache.Stick.X = int32(input.value)
		case C.ABS_HAT0Y:
			dev.nunchuk_cache.Stick.Y = int32(input.value)
		case C.ABS_RX:
			dev.nunchuk_cache.Accel.X = int32(input.value)
		case C.ABS_RY:
			dev.nunchuk_cache.Accel.Y = int32(input.value)
		case C.ABS_RZ:
			dev.nunchuk_cache.Accel.Z = int32(input.value)
		}
	} else if input._type == C.EV_SYN {
		dev.nunchuk_cache.timestamp = cTime(input.time)
		return &dev.nunchuk_cache, nil
	}

	goto try_again
}

func (dev *Device) readClassic() (Event, error) {
	fd := dev.ifs[InterfaceClassicController].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceClassicController)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_KEY {
		if input.value < 0 || input.value > 2 {
			goto try_again
		}

		var key Key
		switch input.code {
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
			goto try_again
		}

		var ev EventClassicControllerKey
		ev.timestamp = cTime(input.time)
		ev.Code = key
		ev.State = KeyState(input.value)
		return &ev, nil
	} else if input._type == C.EV_ABS {
		switch input.code {
		case C.ABS_HAT1X:
			dev.classic_cache.StickLeft.X = int32(input.value)
		case C.ABS_HAT1Y:
			dev.classic_cache.StickLeft.Y = int32(input.value)
		case C.ABS_HAT2X:
			dev.classic_cache.StickRight.X = int32(input.value)
		case C.ABS_HAT2Y:
			dev.classic_cache.StickRight.Y = int32(input.value)
		case C.ABS_HAT3X:
			dev.classic_cache.ShoulderLeft = int32(input.value)
		case C.ABS_HAT3Y:
			dev.classic_cache.ShoulderRight = int32(input.value)
		}
	} else if input._type == C.EV_SYN {
		dev.classic_cache.timestamp = cTime(input.time)
		return &dev.classic_cache, nil
	}

	goto try_again
}

func (dev *Device) readBBoard() (Event, error) {
	fd := dev.ifs[InterfaceBalanceBoard].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceBalanceBoard)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_SYN {
		dev.bboard_cache.timestamp = cTime(input.time)
		return &dev.bboard_cache, nil
	}

	if input._type != C.EV_ABS {
		goto try_again
	}

	switch input.code {
	case C.ABS_HAT0X:
		dev.bboard_cache.Weights[0] = int32(input.value)
	case C.ABS_HAT0Y:
		dev.bboard_cache.Weights[1] = int32(input.value)
	case C.ABS_HAT1X:
		dev.bboard_cache.Weights[2] = int32(input.value)
	case C.ABS_HAT1Y:
		dev.bboard_cache.Weights[3] = int32(input.value)
	}

	goto try_again
}

func (dev *Device) readPro() (Event, error) {
	fd := dev.ifs[InterfaceProController].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceProController)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_KEY {
		if input.value < 0 || input.value > 2 {
			goto try_again
		}

		var key Key
		switch input.code {
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
			goto try_again
		}

		var ev EventProControllerKey
		ev.timestamp = cTime(input.time)
		ev.Code = key
		ev.State = KeyState(input.value)
		return &ev, nil
	} else if input._type == C.EV_ABS {
		switch input.code {
		case C.ABS_X:
			dev.pro_cache.Sticks[0].X = int32(input.value)
		case C.ABS_Y:
			dev.pro_cache.Sticks[0].Y = int32(input.value)
		case C.ABS_RX:
			dev.pro_cache.Sticks[1].X = int32(input.value)
		case C.ABS_RY:
			dev.pro_cache.Sticks[1].Y = int32(input.value)
		}
	} else if input._type == C.EV_SYN {
		dev.pro_cache.timestamp = cTime(input.time)
		return &dev.pro_cache, nil
	}

	goto try_again
}

func (dev *Device) readDrums() (Event, error) {
	fd := dev.ifs[InterfaceDrums].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceDrums)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_KEY {
		if input.value < 0 || input.value > 2 {
			goto try_again
		}

		var key Key
		switch input.code {
		case C.BTN_START:
			key = KeyPlus
		case C.BTN_SELECT:
			key = KeyMinus
		default:
			goto try_again
		}

		var ev EventDrumsKey
		ev.timestamp = cTime(input.time)
		ev.Code = key
		ev.State = KeyState(input.value)
		return &ev, nil
	} else if input._type == C.EV_ABS {
		switch input.code {
		case C.ABS_X:
			dev.drums_cache.Pad.X = int32(input.value)
		case C.ABS_Y:
			dev.drums_cache.Pad.Y = int32(input.value)
		case C.ABS_CYMBAL_LEFT:
			dev.drums_cache.CymbalLeft = int32(input.value)
		case C.ABS_CYMBAL_RIGHT:
			dev.drums_cache.CymbalRight = int32(input.value)
		case C.ABS_TOM_LEFT:
			dev.drums_cache.TomLeft = int32(input.value)
		case C.ABS_TOM_RIGHT:
			dev.drums_cache.TomRight = int32(input.value)
		case C.ABS_TOM_FAR_RIGHT:
			dev.drums_cache.TomFarRight = int32(input.value)
		case C.ABS_BASS:
			dev.drums_cache.Bass = int32(input.value)
		case C.ABS_HI_HAT:
			dev.drums_cache.HiHat = int32(input.value)
		}
	} else if input._type == C.EV_SYN {
		dev.drums_cache.timestamp = cTime(input.time)
		return &dev.drums_cache, nil
	}

	goto try_again
}

func (dev *Device) readGuitar() (Event, error) {
	fd := dev.ifs[InterfaceGuitar].fd
try_again:
	input, err := read_event(fd)
	if err != nil {
		dev.closeInterface(InterfaceGuitar)
		dev.readNodes()
		return &EventWatch{commonEvent{time.Now()}}, nil
	}
	if input == nil {
		return nil, nil
	}

	if input._type == C.EV_KEY {
		if input.value < 0 || input.value > 2 {
			goto try_again
		}

		var key Key
		switch input.code {
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
			goto try_again
		}

		var ev EventGuitarKey
		ev.timestamp = cTime(input.time)
		ev.Code = key
		ev.State = KeyState(input.value)
		return &ev, nil
	} else if input._type == C.EV_ABS {
		switch input.code {
		case C.ABS_X:
			dev.guitar_cache.Stick.X = int32(input.value)
		case C.ABS_Y:
			dev.guitar_cache.Stick.Y = int32(input.value)
		case C.ABS_WHAMMY_BAR:
			dev.guitar_cache.WhammyBar = int32(input.value)
		case C.ABS_FRET_BOARD:
			dev.guitar_cache.FretBar = int32(input.value)
		}
	} else if input._type == C.EV_SYN {
		dev.guitar_cache.timestamp = cTime(input.time)
		return &dev.guitar_cache, nil
	}

	goto try_again
}

func (dev *Device) dispatchEvent(evFd int32, pollEv uint32) (Event, error) {
	if dev.umon != nil && dev.umon.GetFD() == int(evFd) {
		return dev.readUmon(pollEv)
	}
	for name, iff := range dev.ifs {
		if iff.fd.Fd() != uintptr(evFd) {
			continue
		}
		switch name {
		case InterfaceCore:
			return dev.readCore()
		case InterfaceAccel:
			return dev.readAccel()
		case InterfaceIR:
			return dev.readIR()
		case InterfaceMotionPlus:
			return dev.readMP()
		case InterfaceNunchuk:
			return dev.readNunchuck()
		case InterfaceClassicController:
			return dev.readClassic()
		case InterfaceBalanceBoard:
			return dev.readBBoard()
		case InterfaceProController:
			return dev.readPro()
		case InterfaceDrums:
			return dev.readDrums()
		case InterfaceGuitar:
			return dev.readGuitar()
		}
	}
	return nil, nil
}
