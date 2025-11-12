// Package xwiimote has bindings for libxwiimote.
package xwiimote

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
// #include <errno.h>
import "C"
import (
	"os"
	"runtime"
	"unsafe"
)

type Interface int

const (
	// Core interface
	InterfaceCore Interface = C.XWII_IFACE_CORE
	// Accelerometer interface
	InterfaceAccel Interface = C.XWII_IFACE_ACCEL
	// IR interface
	InterfaceIR Interface = C.XWII_IFACE_IR
	// MotionPlus extension interface
	InterfaceMotionPlus Interface = C.XWII_IFACE_MOTION_PLUS
	// Nunchuk extension interface
	InterfaceNunchuk Interface = C.XWII_IFACE_NUNCHUK
	// ClassicController extension interface
	InterfaceClassicController Interface = C.XWII_IFACE_CLASSIC_CONTROLLER
	// BalanceBoard extension interface
	InterfaceBalanceBoard Interface = C.XWII_IFACE_BALANCE_BOARD
	// ProController extension interface
	InterfaceProController Interface = C.XWII_IFACE_PRO_CONTROLLER
	// Drums extension interface
	InterfaceDrums Interface = C.XWII_IFACE_DRUMS
	// Guitar extension interface
	InterfaceGuitar Interface = C.XWII_IFACE_GUITAR
	// Special flag ORed with all valid interfaces
	InterfaceAll Interface = C.XWII_IFACE_ALL
	// Special flag which causes the interfaces to be opened writable
	InterfaceWritable Interface = C.XWII_IFACE_WRITABLE
)

func (dev Interface) Name() string {
	cstr := C.xwii_get_iface_name(C.uint(dev))
	if cstr == nil {
		return "invalid interface"
	}
	return C.GoString(cstr)
}

type Led uint

const (
	Led1 Led = C.XWII_LED1
	Led2 Led = C.XWII_LED2
	Led3 Led = C.XWII_LED3
	Led4 Led = C.XWII_LED4
)

type Device struct {
	cptr *C.struct_xwii_iface
}

// Create new device object from syspath path
//
// @param[out] dev Pointer to new opaque device is stored here
// @param[in] syspath Sysfs path to root device node
//
// Creates a new device object. No interfaces on the device are opened by
// default. @p syspath must be a valid path to a wiimote device, either
// retrieved via a @ref monitor "monitor object" or via udev. It must point to
// the hid device, which is normally /sys/bus/hid/devices/[dev].
//
// If this function fails, @p dev is not touched at all (and not cleared!). A
// new object always has an initial ref-count of 1.
//
// @returns 0 on success, negative error code on failure
func NewDevice(syspath string) (*Device, error) {
	var dev Device
	csyspath := C.CString(syspath)
	defer C.free(unsafe.Pointer(csyspath))

	ret := C.xwii_iface_new(&dev.cptr, csyspath)
	if ret != 0 {
		return nil, cError(ret)
	}
	runtime.SetFinalizer(&dev, func(i *Device) {
		i.Free()
	})

	return &dev, cError(ret)
}

func (dev *Device) Free() {
	if dev.cptr == nil {
		return
	}
	runtime.SetFinalizer(&dev, nil)
	C.xwii_iface_unref(dev.cptr)
	dev.cptr = nil
}

// Return device syspath
//
// This returns the sysfs path of the underlying device. It is not neccesarily
// the same as the one during xwii_iface_new(). However, it is guaranteed to
// point at the same device (symlinks may be resolved).
//
// @returns NULL on failure, otherwise a constant device syspath is returned.
func (dev *Device) GetSyspath() string {
	cstr := C.xwii_iface_get_syspath(dev.cptr)
	if cstr == nil {
		return ""
	}
	return C.GoString(cstr)
}

// Return file-descriptor
//
// Return the file-descriptor used by this device. If multiple file-descriptors
// are used internally, they are multi-plexed through an epoll descriptor.
// Therefore, this always returns the same single file-descriptor. You need to
// watch this for readable-events (POLLIN/EPOLLIN) and call
// xwii_iface_dispatch() whenever it is readable.
//
// This function always returns a valid file-descriptor.
func (dev *Device) GetFD() int {
	return int(C.xwii_iface_get_fd(dev.cptr))
}

// Watch device for hotplug events
//
// @param[in] watch Whether to watch for hotplug events or not
//
// Toggle whether hotplug events should be reported or not. By default, no
// hotplug events are reported so this is off.
//
// Note that this requires a separate udev-monitor for each device. Therefore,
// if your application uses its own udev-monitor, you should instead integrate
// the hotplug-detection into your udev-monitor.
//
// @returns 0 on success, negative error code on failure
func (dev *Device) Watch(hotplug bool) error {
	ret := C.xwii_iface_watch(dev.cptr, C.bool(hotplug))
	return cError(ret)
}

// Open interfaces on this device
//
// @param[in] ifaces Bitmask of interfaces of type enum xwii_iface_type
//
// Open all the requested interfaces. If @ref XWII_IFACE_WRITABLE is also set,
// the interfaces are opened with write-access. Note that interfaces that are
// already opened are ignored and not touched.
// If _any_ interface fails to open, this function still tries to open the other
// requested interfaces and then returns the error afterwards. Hence, if this
// function fails, you should use xwii_iface_opened() to get a bitmask of opened
// interfaces and see which failed (if that is of interest).
//
// Note that interfaces may be closed automatically during runtime if the
// kernel removes the interface or on error conditions. You always get an
// @ref XWII_EVENT_WATCH event which you should react on. This is returned
// regardless whether xwii_iface_watch() was enabled or not.
//
// @returns 0 on success, negative error code on failure.
func (dev *Device) Open(ifaces Interface) error {
	ret := C.xwii_iface_open(dev.cptr, C.uint(ifaces))
	return cError(ret)
}

// Close interfaces on this device
//
// @param[in] ifaces Bitmask of interfaces of type enum xwii_iface_type
//
// Close the requested interfaces. This never fails.
func (dev *Device) Close(ifaces Interface) {
	C.xwii_iface_close(dev.cptr, C.uint(ifaces))
}

// Return bitmask of opened interfaces
//
// Returns a bitmask of opened interfaces. Interfaces may be closed due to
// error-conditions at any time. However, interfaces are never opened
// automatically.
//
// You will get notified whenever this bitmask changes, except on explicit
// calls to xwii_iface_open() and xwii_iface_close(). See the
// @ref XWII_EVENT_WATCH event for more information.
func (dev *Device) Opened() Interface {
	ret := C.xwii_iface_opened(dev.cptr)
	return Interface(ret)
}

// Return bitmask of available interfaces
//
// Return a bitmask of available devices. These devices can be opened and are
// guaranteed to be present on the hardware at this time. If you watch your
// device for hotplug events (see xwii_iface_watch()) you will get notified
// whenever this bitmask changes. See the @ref XWII_EVENT_WATCH event for more
// information.
func (dev *Device) Available() Interface {
	ret := C.xwii_iface_available(dev.cptr)
	return Interface(ret)
}

// Read incoming event-queue
//
// @param[out] ev Pointer where to store a new event or NULL
// @param[in] size Size of @p ev if @p ev is non-NULL
//
// You should call this whenever the file-descriptor returned by
// xwii_iface_get_fd() is reported as being readable. This function will perform
// all non-blocking outstanding tasks and then return.
//
// This function always performs any background tasks and outgoing event-writes
// if they don't block. It returns an error if they fail.
// If @p ev is NULL, this function returns 0 on success after this has been
// done.
//
// If @p ev is non-NULL, this function then tries to read a single incoming
// event. If no event is available, it returns -EAGAIN and you should watch the
// file-desciptor again until it is readable. Otherwise, you should call this
// function in a row as long as it returns 0. It stores the event in @p ev which
// you can then handle in your application.
//
// This function is the successor or xwii_iface_poll(). It takes an additional
// @p size argument to provide backwards compatibility.
//
// @returns 0 on success, -EAGAIN if no event can be read and @p ev is non-NULL
// and a negative error-code on failure
func (dev *Device) Dispatch() (Event, error) {
	var ev C.struct_xwii_event
	ret := C.xwii_iface_dispatch(dev.cptr, &ev, C.size_t(unsafe.Sizeof(ev)))
	if ret == -C.EAGAIN {
		return nil, nil
	}
	if ret != 0 {
		return nil, cError(ret)
	}

	return makeEvent(ev)
}

func makeEvent(ev C.struct_xwii_event) (Event, error) {
	switch ev._type {
	case C.XWII_EVENT_KEY:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventKey{timestamp: cTime(ev.time), Code: Key(payload.code), State: KeyState(payload.state)}, nil

	case C.XWII_EVENT_ACCEL:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventAccel{timestamp: cTime(ev.time)}
		rev.Accel.X = int32(payload[0].x)
		rev.Accel.Y = int32(payload[0].y)
		rev.Accel.Z = int32(payload[0].z)
		return rev, nil

	case C.XWII_EVENT_IR:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventIR{timestamp: cTime(ev.time)}
		for i := range rev.Slots {
			rev.Slots[i].X = int32(payload[i].x)
			rev.Slots[i].Y = int32(payload[i].y)
		}
		return rev, nil

	case C.XWII_EVENT_BALANCE_BOARD:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventBalanceBoard{timestamp: cTime(ev.time)}
		for i := range rev.Weights {
			rev.Weights[i] = int32(payload[i].x)
		}
		return rev, nil

	case C.XWII_EVENT_MOTION_PLUS:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventMotionPlus{timestamp: cTime(ev.time)}
		for i := range rev.Speed {
			rev.Speed[i].X = int32(payload[i].x)
			rev.Speed[i].Y = int32(payload[i].y)
			rev.Speed[i].Z = int32(payload[i].z)
		}
		return rev, nil

	case C.XWII_EVENT_PRO_CONTROLLER_KEY:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventProControllerKey{timestamp: cTime(ev.time), Code: Key(payload.code), State: KeyState(payload.state)}, nil

	case C.XWII_EVENT_PRO_CONTROLLER_MOVE:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventProControllerMove{timestamp: cTime(ev.time)}
		for i := range rev.Sticks {
			rev.Sticks[i].X = int32(payload[i].x)
			rev.Sticks[i].Y = int32(payload[i].y)
		}
		return rev, nil

	case C.XWII_EVENT_WATCH:
		return &EventWatch{timestamp: cTime(ev.time)}, nil

	case C.XWII_EVENT_CLASSIC_CONTROLLER_KEY:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventClassicControllerKey{timestamp: cTime(ev.time), Code: Key(payload.code), State: KeyState(payload.state)}, nil

	case C.XWII_EVENT_CLASSIC_CONTROLLER_MOVE:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventClassicControllerMove{timestamp: cTime(ev.time)}
		for i := range rev.Sticks {
			rev.Sticks[i].X = int32(payload[i].x)
			rev.Sticks[i].Y = int32(payload[i].y)
		}
		return rev, nil

	case C.XWII_EVENT_NUNCHUK_KEY:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventNunchukKey{timestamp: cTime(ev.time), Code: Key(payload.code), State: KeyState(payload.state)}, nil

	case C.XWII_EVENT_NUNCHUK_MOVE:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventNunchukMove{timestamp: cTime(ev.time)}
		rev.Stick.X = int32(payload[0].x)
		rev.Stick.Y = int32(payload[0].y)
		rev.Accel.X = int32(payload[1].x)
		rev.Accel.Y = int32(payload[1].y)
		rev.Accel.Z = int32(payload[1].z)
		return rev, nil

	case C.XWII_EVENT_DRUMS_KEY:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventDrumsKey{timestamp: cTime(ev.time), Code: Key(payload.code), State: KeyState(payload.state)}, nil

	case C.XWII_EVENT_DRUMS_MOVE:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventDrumsMove{timestamp: cTime(ev.time)}
		rev.Pad.X = int32(payload[C.XWII_DRUMS_ABS_PAD].x)
		rev.Pad.Y = int32(payload[C.XWII_DRUMS_ABS_PAD].y)
		rev.CymbalLeft = int32(payload[C.XWII_DRUMS_ABS_CYMBAL_LEFT].x)
		rev.CymbalRight = int32(payload[C.XWII_DRUMS_ABS_CYMBAL_RIGHT].x)
		rev.TomLeft = int32(payload[C.XWII_DRUMS_ABS_TOM_LEFT].x)
		rev.TomRight = int32(payload[C.XWII_DRUMS_ABS_TOM_RIGHT].x)
		rev.TomFarRight = int32(payload[C.XWII_DRUMS_ABS_TOM_FAR_RIGHT].x)
		rev.Bass = int32(payload[C.XWII_DRUMS_ABS_BASS].x)
		rev.HiHat = int32(payload[C.XWII_DRUMS_ABS_HI_HAT].x)
		return rev, nil

	case C.XWII_EVENT_GUITAR_KEY:
		payload := (*C.struct_xwii_event_key)(unsafe.Pointer(&ev.v))
		return &EventGuitarKey{timestamp: cTime(ev.time), Code: Key(payload.code), State: KeyState(payload.state)}, nil

	case C.XWII_EVENT_GUITAR_MOVE:
		payload := (*[C.XWII_ABS_NUM]C.struct_xwii_event_abs)(unsafe.Pointer(&ev.v))
		rev := &EventGuitarMove{timestamp: cTime(ev.time)}
		rev.Stick.X = int32(payload[0].x)
		rev.Stick.Y = int32(payload[0].y)
		rev.WhammyBar = int32(payload[1].x)
		rev.FretBar = int32(payload[2].z)
		return rev, nil

	case C.XWII_EVENT_GONE:
		return &EventGone{timestamp: cTime(ev.time)}, nil
	}
	return nil, os.ErrInvalid
}

// Make background work
//
// You should call this whenever the file-descriptor returned by
// xwii_iface_get_fd() is reported as being readable. This function will perform
// all non-blocking outstanding tasks and then return.
//
// This function always performs any background tasks and outgoing event-writes
// if they don't block. It returns an error if they fail.
// If @p ev is NULL, this function returns 0 on success after this has been
// done.
//
// If @p ev is non-NULL, this function then tries to read a single incoming
// event. If no event is available, it returns -EAGAIN and you should watch the
// file-desciptor again until it is readable. Otherwise, you should call this
// function in a row as long as it returns 0. It stores the event in @p ev which
// you can then handle in your application.
//
// This function is the successor or xwii_iface_poll(). It takes an additional
// @p size argument to provide backwards compatibility.
//
// @returns 0 on success, -EAGAIN if no event can be read and @p ev is non-NULL
// and a negative error-code on failure
func (dev *Device) Work() error {
	ret := C.xwii_iface_dispatch(dev.cptr, nil, 0)
	return cError(ret)
}

// Toggle rumble motor
//
// @param[in] on New rumble motor state
//
// Toggle the rumble motor. This requires the core-interface to be opened in
// writable mode.
//
// @returns 0 on success, negative error code on failure.
func (dev *Device) Rumble(state bool) error {
	ret := C.xwii_iface_watch(dev.cptr, C.bool(state))
	return cError(ret)
}

// Read LED state
//
// @param[in] led LED constant defined in enum xwii_led
// @param[out] state Pointer where state should be written to
//
// Reads the current LED state of the given LED. @p state will be either true or
// false depending on whether the LED is on or off.
//
// LEDs are a static interface that does not have to be opened first.
//
// @returns 0 on success, negative error code on failure
func (dev *Device) GetLED(led Led) (bool, error) {
	var state C.bool
	ret := C.xwii_iface_get_led(dev.cptr, C.uint(led), &state)
	return bool(state), cError(ret)
}

// Set LED state
//
// @param[in] led LED constant defined in enum xwii_led
// @param[in] state State to set on the LED
//
// Changes the current LED state of the given LED. This has immediate effect.
//
// LEDs are a static interface that does not have to be opened first.
//
// @returns 0 on success, negative error code on failure
func (dev *Device) SetLED(led Led, state bool) error {
	ret := C.xwii_iface_set_led(dev.cptr, C.uint(led), C.bool(state))
	return cError(ret)
}

// Read battery state
//
// @param[out] capacity Pointer where state should be written to
//
// Reads the current battery capacity and write it into @p capacity. This is
// a value between 0 and 100, which describes the current capacity in per-cent.
//
// Batteries are a static interface that does not have to be opened first.
//
// @returns 0 on success, negative error code on failure
func (dev *Device) GetBattery() (uint8, error) {
	var state C.uint8_t
	ret := C.xwii_iface_get_battery(dev.cptr, &state)
	return uint8(state), cError(ret)
}

// Read device type
//
// @param[out] devtype Pointer where the device type should be stored
//
// Reads the current device-type, allocates a string and stores a pointer to
// the string in @p devtype. You must free it via free() after you are done.
//
// This is a static interface that does not have to be opened first.
//
// @returns 0 on success, negative error code on failure
func (dev *Device) GetDevType() (string, error) {
	var devtype *C.char
	ret := C.xwii_iface_get_devtype(dev.cptr, &devtype)
	if ret != 0 {
		return "", cError(ret)
	}
	return cStringCopy(devtype), nil
}

// Read extension type
//
// @param[out] extension Pointer where the extension type should be stored
//
// Reads the current extension type, allocates a string and stores a pointer
// to the string in @p extension. You must free it via free() after you are
// done.
//
// This is a static interface that does not have to be opened first.
//
// @returns 0 on success, negative error code on failure
func (dev *Device) GetExtension() (string, error) {
	var ext *C.char
	ret := C.xwii_iface_get_extension(dev.cptr, &ext)
	if ret != 0 {
		return "", cError(ret)
	}
	return cStringCopy(ext), nil
}

// Set MP normalization and calibration
//
// @param[in] x x-value to use or 0
// @param[in] y y-value to use or 0
// @param[in] z z-value to use or 0
// @param[in] factor factor-value to use or 0
//
// Set MP-normalization and calibration values. The Motion-Plus sensor is very
// sensitive and may return really crappy values. This interfaces allows to
// apply 3 absolute offsets x, y and z which are subtracted from any MP data
// before it is returned to the application. That is, if you set these values
// to 0, this has no effect (which is also the initial state).
//
// The calibration factor @p factor is used to perform runtime calibration. If
// it is 0 (the initial state), no runtime calibration is performed. Otherwise,
// the factor is used to re-calibrate the zero-point of MP data depending on MP
// input. This is an angoing calibration which modifies the internal state of
// the x, y and z values.
func (dev *Device) SetMPNormalization(x, y, z, factor int32) {
	C.xwii_iface_set_mp_normalization(dev.cptr, C.int32_t(x), C.int32_t(y), C.int32_t(z), C.int32_t(factor))
}

// Read MP normalization and calibration
//
// @param[out] x Pointer where to store x-value or NULL
// @param[out] y Pointer where to store y-value or NULL
// @param[out] z Pointer where to store z-value or NULL
// @param[out] factor Pointer where to store factor-value or NULL
//
// Reads the MP normalization and calibration values. Please see
// xwii_iface_set_mp_normalization() how this is handled.
//
// Note that if the calibration factor is not 0, the normalization values may
// change depending on incoming MP data. Therefore, the data read via this
// function may differ from the values that you wrote to previously. However,
// apart from applied calibration, these value are the same as were set
// previously via xwii_iface_set_mp_normalization() and you can feed them back
// in later.
func (dev *Device) GetMPNormalization() (x, y, z, factor int32) {
	C.xwii_iface_get_mp_normalization(dev.cptr, (*C.int32_t)(&x), (*C.int32_t)(&y), (*C.int32_t)(&z), (*C.int32_t)(&factor))
	return
}
