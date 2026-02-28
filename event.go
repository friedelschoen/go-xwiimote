package wiimote

import "time"

// Key Event Identifiers
//
// For each key found on a supported device, a separate key identifier is
// defined. Note that a device may have a specific key (for instance: HOME) on
// the main device and on an extension device. An application can detect which
// key was pressed examining the event-type field.
// Some devices report common keys as both, extension and core events. In this
// case the kernel is required to filter these and you should report it as a
// bug. A single physical key-press should never be reported twice, even on two
// different features.
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

// Vec2 represents a 2D point or vector to X and Y, may be interpreted different depending on the event .
type Vec2 struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

// Vec3 represents a 3D point or vector to X, Y and Z, may be interpreted different depending on the event.
type Vec3 struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
	Z int32 `json:"z"`
}

// Event feature describes an event fired by Device.Dispatch(),
// consider using a type-switch to retrieve the specific event type and data
type Event interface {
	Feature() Feature
	Timestamp() time.Time
}

// EventKey is fired whenever a key is pressed or released. Valid
// key-events include all the events reported by the core-feature,
// which is normally only LEFT, RIGHT, UP, DOWN, A, B, PLUS, MINUS,
// HOME, ONE, TWO.
type EventKey struct {
	Event
	Code    Key  `json:"code"`
	Pressed bool `json:"pressed"`
}

// EventAccel provides accelerometer data.
// Note that the accelerometer reports acceleration data, not speed
// data!
type EventAccel struct {
	Event
	Accel Vec3 `json:"accel"`
}

// EventIR provides IR-camera events. The camera can track up two four IR
// sources. As long as a single source is tracked, it stays at it's
// pre-allocated slot.
//
// Use IRSlot.Valid() to see whether a specific slot is
// currently valid or whether it currently doesn't track any IR source.
type EventIR struct {
	Event
	Slots [4]IRSlot `json:"slots"`
}

// IRSlot describes Infra-Red Tracking on a WiiMote
type IRSlot struct {
	Vec2
	Size uint8
}

// Valid returns wether this slot holds a valid source. If not it has no track and is considered disabled.
func (slot IRSlot) Valid() bool {
	return slot.X != 1023 || slot.Y != 1023
}

// EventBalanceBoard provides balance-board weight data. Four sensors report weight-data
// for each of the four edges of the board.
type EventBalanceBoard struct {
	Event
	Weights [4]int32 `json:"weights"`
}

// EventMotionPlus provides gyroscope events. These describe rotational speed, not
// acceleration, of the motion-plus extension.
type EventMotionPlus struct {
	Event
	Speed Vec3 `json:"speed"`
}

// EventProControllerKey provides button events of the pro-controller
// and are reported via this feature
// and not via the core-feature (which only reports core-buttons).
// Valid buttons include: LEFT, RIGHT, UP, DOWN, PLUS, MINUS, HOME, X,
// Y, A, B, TR, TL, ZR, ZL, THUMBL, THUMBR.
// Payload type is struct wii_event_key.
type EventProControllerKey struct {
	EventKey
}

// EventProControllerMove provides movement of analog sticks on the pro-controller and is
// reported via this event.
type EventProControllerMove struct {
	Event
	Sticks [2]Vec2 `json:"sticks"`
}

// EventWatch is sent whenever an extension was hotplugged (plugged or
// unplugged), a device-detection finished or some other static data
// changed which cannot be monitored separately.
// An application should check what changed by examining the device is
// testing whether all required features are still available.
// Non-hotplug aware devices may discard this event.
//
// This is only returned if you explicitly watched for hotplug events.
// See Device.Watch().
//
// This event is also returned if an feature is closed because the
// kernel closed our file-descriptor (for whatever reason). This is
// returned regardless whether you watch for hotplug events or not.
type EventWatch struct {
	Event
}

// EventClassicControllerKey provides Classic Controller key events.
// Button events of the classic controller are reported via this
// feature and not via the core-feature (which only reports
// core-buttons).
// Valid buttons include: LEFT, RIGHT, UP, DOWN, PLUS, MINUS, HOME, X,
// Y, A, B, TR, TL, ZR, ZL.
type EventClassicControllerKey struct {
	EventKey
}

// EventClassicControllerMove provides Classic Controller movement events.
// Movement of analog sticks are reported via this event. The payload
// is a struct wii_event_abs and the first two array elements contain
// the absolute x and y position of both analog sticks.
// The x value of the third array element contains the absolute position
// of the TL trigger. The y value contains the absolute position for the
// TR trigger. Note that many classic controllers do not have analog
// TL/TR triggers, in which case these read 0 or MAX (63). The digital
// TL/TR buttons are always reported correctly.
type EventClassicControllerMove struct {
	Event
	StickLeft     Vec2  `json:"stick_left"`
	StickRight    Vec2  `json:"stick_right"`
	ShoulderLeft  int32 `json:"shoulder_left"`
	ShoulderRight int32 `json:"shoulder_right"`
}

// EventNunchukKey provides Nunchuk key events.
// Button events of the nunchuk controller are reported via this
// feature and not via the core-feature (which only reports
// core-buttons).
// Valid buttons include: C, Z
type EventNunchukKey struct {
	EventKey
}

// EventNunchukMove provides Nunchuk movement events.
// Movement events of the nunchuk controller are reported via this
// feature. Payload is of type struct wii_event_abs. The first array
// element contains the x/y positions of the analog stick. The second
// array element contains the accelerometer information.
type EventNunchukMove struct {
	Event
	Stick Vec2 `json:"stick"`
	Accel Vec3 `json:"accel"`
}

// EventDrumsKey provides Drums key events.
// Button events for drums controllers. Valid buttons are PLUS and MINUS
// for the +/- buttons on the center-bar.
type EventDrumsKey struct {
	EventKey
}

// EventDrumsMove provides Drums movement event
// Movement and pressure events for drums controllers.
type EventDrumsMove struct {
	Event
	Pad         Vec2  `json:"pad"`
	CymbalLeft  int32 `json:"cymbal_left"`
	CymbalRight int32 `json:"cymbal_right"`
	TomLeft     int32 `json:"tom_left"`
	TomRight    int32 `json:"tom_right"`
	TomFarRight int32 `json:"tom_far_right"`
	Bass        int32 `json:"bass"`
	HiHat       int32 `json:"hihat"`
}

// EventGuitarKey provides Guitar key events
// Button events for guitar controllers. Valid buttons are HOME and PLUS
// for the StarPower/Home button and the + button. Furthermore, you get
// FRET_FAR_UP, FRET_UP, FRET_MID, FRET_LOW, FRET_FAR_LOW for fret
// activity and STRUM_BAR_UP and STRUM_BAR_LOW for the strum bar.
type EventGuitarKey struct {
	EventKey
}

// EventGuitarMove provides Guitar movement events.
// Movement information for guitar controllers.
type EventGuitarMove struct {
	Event
	Stick     Vec2  `json:"stick"`
	WhammyBar int32 `json:"whammy_bar"`
	FretBar   int32 `json:"fret_bar"`
}

// EventFeature is provided when a feature is added.
type EventFeature struct {
	Event
	Kind    FeatureKind
	Removed bool
}

// EventGone provides Removal Event.
// This event is sent whenever the device was removed. No payload is provided.
// Non-hotplug aware applications may discard this event.
// Optionally Feature can be set, if nil the whole device is removed.
type EventGone struct {
	Event
}
