package xwiimote

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
import "C"
import (
	"fmt"
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
	KeyLeft  Key = C.XWII_KEY_LEFT
	KeyRight Key = C.XWII_KEY_RIGHT
	KeyUp    Key = C.XWII_KEY_UP
	KeyDown  Key = C.XWII_KEY_DOWN
	KeyA     Key = C.XWII_KEY_A
	KeyB     Key = C.XWII_KEY_B
	KeyPlus  Key = C.XWII_KEY_PLUS
	KeyMinus Key = C.XWII_KEY_MINUS
	KeyHome  Key = C.XWII_KEY_HOME
	KeyOne   Key = C.XWII_KEY_ONE
	KeyTwo   Key = C.XWII_KEY_TWO
	KeyX     Key = C.XWII_KEY_X
	KeyY     Key = C.XWII_KEY_Y
	KeyTL    Key = C.XWII_KEY_TL
	KeyTR    Key = C.XWII_KEY_TR
	KeyZL    Key = C.XWII_KEY_ZL
	KeyZR    Key = C.XWII_KEY_ZR

	// Left thumb button
	//
	// This is reported if the left analog stick is pressed. Not all analog
	// sticks support this. The Wii-U Pro Controller is one of few devices
	// that report this event.
	KeyThumbL Key = C.XWII_KEY_THUMBL

	// Right thumb button
	//
	// This is reported if the right analog stick is pressed. Not all analog
	// sticks support this. The Wii-U Pro Controller is one of few devices
	// that report this event.
	KeyThumbR Key = C.XWII_KEY_THUMBR

	// Extra C button
	//
	// This button is not part of the standard action pad but reported by
	// extension controllers like the Nunchuk. It is supposed to extend the
	// standard A and B buttons.
	KeyC Key = C.XWII_KEY_C

	// Extra Z button
	//
	// This button is not part of the standard action pad but reported by
	// extension controllers like the Nunchuk. It is supposed to extend the
	// standard X and Y buttons.
	KeyZ Key = C.XWII_KEY_Z

	// Guitar Strum-bar-up event
	//
	// Emitted by guitars if the strum-bar is moved up.
	KeyStrumBarUp Key = C.XWII_KEY_STRUM_BAR_UP

	// Guitar Strum-bar-down event
	//
	// Emitted by guitars if the strum-bar is moved down.
	KeyStrumBarDown Key = C.XWII_KEY_STRUM_BAR_DOWN

	// Guitar Fret-Far-Up event
	//
	// Emitted by guitars if the upper-most fret-bar is pressed.
	KeyFretFarUp Key = C.XWII_KEY_FRET_FAR_UP

	// Guitar Fret-Up event
	//
	// Emitted by guitars if the second-upper fret-bar is pressed.
	KeyFretUp Key = C.XWII_KEY_FRET_UP

	// Guitar Fret-Mid event
	//
	// Emitted by guitars if the mid fret-bar is pressed.
	KeyFretMid Key = C.XWII_KEY_FRET_MID

	// Guitar Fret-Low event
	//
	// Emitted by guitars if the second-lowest fret-bar is pressed.
	KeyFretLow Key = C.XWII_KEY_FRET_LOW

	// Guitar Fret-Far-Low event
	//
	// Emitted by guitars if the lower-most fret-bar is pressed.
	KeyFretFarLow Key = C.XWII_KEY_FRET_FAR_LOW
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

// Vec2 represents a 2D point or vector to X and Y, may be interpreted otherwise depending on the event .
type Vec2 struct{ X, Y int32 }

// Vec3 represents a 3D point or vector to X, Y and Z, may be interpreted otherwise depending on the event.
type Vec3 struct{ X, Y, Z int32 }

// EventType describes the type of event as an integer.
type EventType uint

const (
	EventTypeKey                   EventType = C.XWII_EVENT_KEY
	EventTypeAccel                 EventType = C.XWII_EVENT_ACCEL
	EventTypeIR                    EventType = C.XWII_EVENT_IR
	EventTypeBalanceBoard          EventType = C.XWII_EVENT_BALANCE_BOARD
	EventTypeMotionPlus            EventType = C.XWII_EVENT_MOTION_PLUS
	EventTypeProControllerKey      EventType = C.XWII_EVENT_PRO_CONTROLLER_KEY
	EventTypeProControllerMove     EventType = C.XWII_EVENT_PRO_CONTROLLER_MOVE
	EventTypeWatch                 EventType = C.XWII_EVENT_WATCH
	EventTypeClassicControllerKey  EventType = C.XWII_EVENT_CLASSIC_CONTROLLER_KEY
	EventTypeClassicControllerMove EventType = C.XWII_EVENT_CLASSIC_CONTROLLER_MOVE
	EventTypeNunchukKey            EventType = C.XWII_EVENT_NUNCHUK_KEY
	EventTypeNunchukMove           EventType = C.XWII_EVENT_NUNCHUK_MOVE
	EventTypeDrumsKey              EventType = C.XWII_EVENT_DRUMS_KEY
	EventTypeDrumsMove             EventType = C.XWII_EVENT_DRUMS_MOVE
	EventTypeGuitarKey             EventType = C.XWII_EVENT_GUITAR_KEY
	EventTypeGuitarMove            EventType = C.XWII_EVENT_GUITAR_MOVE
	EventTypeGone                  EventType = C.XWII_EVENT_GONE
)

// Event interface describes an event fired by Device.Dispatch(),
// consider using a type-switch to retrieve the specific event type and data
type Event interface {
	fmt.Stringer
	Timestamp() time.Time
	Type() EventType
}

type commonEvent struct {
	typ       EventType
	timestamp time.Time
}

func (evt *commonEvent) Timestamp() time.Time {
	return evt.timestamp
}

func (evt *commonEvent) Type() EventType {
	return evt.typ
}

func (evt *commonEvent) String() string {
	return fmt.Sprintf("xwiimote-event[type=%v, timestamp=%v]", evt.typ, evt.timestamp)
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
	Speed [3]Vec3
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

func vec2event(payload C.struct_xwii_event_abs) Vec2 {
	return Vec2{
		X: int32(payload.x),
		Y: int32(payload.y),
	}
}

func vec3event(payload C.struct_xwii_event_abs) Vec3 {
	return Vec3{
		X: int32(payload.x),
		Y: int32(payload.y),
		Z: int32(payload.z),
	}
}

func makeEvent(ev C.struct_xwii_event) Event {
	common := commonEvent{typ: EventType(ev._type), timestamp: cTime(ev.time)}

	switch common.Type() {
	case EventTypeKey:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventKey{commonEvent: common, Code: Key(payload.code), State: KeyState(payload.state)}

	case EventTypeAccel:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventAccel{commonEvent: common}
		rev.Accel = vec3event(payload[0])
		return rev

	case EventTypeIR:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventIR{commonEvent: common}
		for i := range rev.Slots {
			rev.Slots[i] = IRSlot{vec2event(payload[i])}
		}
		return rev

	case EventTypeBalanceBoard:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventBalanceBoard{commonEvent: common}
		for i := range rev.Weights {
			rev.Weights[i] = int32(payload[i].x)
		}
		return rev

	case EventTypeMotionPlus:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventMotionPlus{commonEvent: common}
		for i := range rev.Speed {
			rev.Speed[i] = vec3event(payload[i])
		}
		return rev

	case EventTypeProControllerKey:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventProControllerKey{commonEvent: common, Code: Key(payload.code), State: KeyState(payload.state)}

	case EventTypeProControllerMove:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventProControllerMove{commonEvent: common}
		for i := range rev.Sticks {
			rev.Sticks[i] = vec2event(payload[i])
		}
		return rev

	case EventTypeWatch:
		return &EventWatch{commonEvent: common}

	case EventTypeClassicControllerKey:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventClassicControllerKey{commonEvent: common, Code: Key(payload.code), State: KeyState(payload.state)}

	case EventTypeClassicControllerMove:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventClassicControllerMove{commonEvent: common}
		rev.StickLeft = vec2event(payload[0])
		rev.StickRight = vec2event(payload[1])
		rev.ShoulderLeft = int32(payload[2].x)
		rev.ShoulderRight = int32(payload[2].y)
		return rev

	case EventTypeNunchukKey:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventNunchukKey{commonEvent: common, Code: Key(payload.code), State: KeyState(payload.state)}

	case EventTypeNunchukMove:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventNunchukMove{commonEvent: common}
		rev.Stick = vec2event(payload[0])
		rev.Accel = vec3event(payload[1])
		return rev

	case EventTypeDrumsKey:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventDrumsKey{commonEvent: common, Code: Key(payload.code), State: KeyState(payload.state)}

	case EventTypeDrumsMove:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventDrumsMove{commonEvent: common}
		rev.Pad = vec2event(payload[C.XWII_DRUMS_ABS_PAD])
		rev.CymbalLeft = int32(payload[C.XWII_DRUMS_ABS_CYMBAL_LEFT].x)
		rev.CymbalRight = int32(payload[C.XWII_DRUMS_ABS_CYMBAL_RIGHT].x)
		rev.TomLeft = int32(payload[C.XWII_DRUMS_ABS_TOM_LEFT].x)
		rev.TomRight = int32(payload[C.XWII_DRUMS_ABS_TOM_RIGHT].x)
		rev.TomFarRight = int32(payload[C.XWII_DRUMS_ABS_TOM_FAR_RIGHT].x)
		rev.Bass = int32(payload[C.XWII_DRUMS_ABS_BASS].x)
		rev.HiHat = int32(payload[C.XWII_DRUMS_ABS_HI_HAT].x)
		return rev

	case EventTypeGuitarKey:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventGuitarKey{commonEvent: common, Code: Key(payload.code), State: KeyState(payload.state)}

	case EventTypeGuitarMove:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventGuitarMove{commonEvent: common}
		rev.Stick = vec2event(payload[0])
		rev.WhammyBar = int32(payload[1].x)
		rev.FretBar = int32(payload[2].x)
		return rev

	case EventTypeGone:
		return &EventGone{commonEvent: common}
	}
	return nil
}
