// Package xwiimote has bindings for libxwiimote, a library to read and control inputs on a Nintendo WiiMote and accessories.
package xwiimote

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
// #include <errno.h>
import "C"
import (
	"os"
	"runtime"
	"strings"
	"unsafe"
)

type InterfaceType int

const (
	// Core interface
	InterfaceCore InterfaceType = C.XWII_IFACE_CORE
	// Accelerometer interface
	InterfaceAccel InterfaceType = C.XWII_IFACE_ACCEL
	// IR interface
	InterfaceIR InterfaceType = C.XWII_IFACE_IR
	// MotionPlus extension interface
	InterfaceMotionPlus InterfaceType = C.XWII_IFACE_MOTION_PLUS
	// Nunchuk extension interface
	InterfaceNunchuk InterfaceType = C.XWII_IFACE_NUNCHUK
	// ClassicController extension interface
	InterfaceClassicController InterfaceType = C.XWII_IFACE_CLASSIC_CONTROLLER
	// BalanceBoard extension interface
	InterfaceBalanceBoard InterfaceType = C.XWII_IFACE_BALANCE_BOARD
	// ProController extension interface
	InterfaceProController InterfaceType = C.XWII_IFACE_PRO_CONTROLLER
	// Drums extension interface
	InterfaceDrums InterfaceType = C.XWII_IFACE_DRUMS
	// Guitar extension interface
	InterfaceGuitar InterfaceType = C.XWII_IFACE_GUITAR
	// Special flag ORed with all valid interfaces
	InterfaceAll InterfaceType = C.XWII_IFACE_ALL
	// Special flag which causes the interfaces to be opened writable
	InterfaceWritable InterfaceType = C.XWII_IFACE_WRITABLE
)

// Name returns the original name of that interface.
func (dev InterfaceType) Name() string {
	cstr := C.xwii_get_iface_name(C.uint(dev))
	if cstr == nil {
		return "invalid interface"
	}
	return C.GoString(cstr)
}

// Led described a Led of an device. The leds are counted left-to-right and can be OR'ed together.
type Led uint

const (
	Led1 Led = 1 << iota
	Led2
	Led3
	Led4
)

// Device describes the communication with a single device. That is, you
// create one for each device you use. All sub-interfaces are opened on this
// object.
type Device struct {
	poller[Event]
	cptr *C.struct_xwii_iface
}

// NewDevice creates a new device object. No interfaces on the device are opened by
// default.
//
// syspath must be a valid path to a wiimote device, either
// retrieved via a Monitor, an Enumerator or via udev directly. It must point to
// the hid device, which is normally /sys/bus/hid/devices/[dev].
//
// The object and underlying structure is freed automatically by default.
func NewDevice(syspath string) (*Device, error) {
	dev := new(Device)
	dev.poller = newPoller(dev)
	csyspath := C.CString(syspath)
	defer C.free(unsafe.Pointer(csyspath))

	ret := C.xwii_iface_new(&dev.cptr, csyspath)
	if ret != 0 {
		return nil, cError(ret)
	}
	runtime.SetFinalizer(dev, func(i *Device) {
		i.Free()
	})

	return dev, cError(ret)
}

// Free unreferences the devices and frees the underlying structure.
// Calling Free is not mandatory and is done automatically by default.
func (dev *Device) Free() {
	if dev.cptr == nil {
		return
	}
	runtime.SetFinalizer(dev, nil)
	C.xwii_iface_unref(dev.cptr)
	dev.cptr = nil
}

// GetSyspath returns the sysfs path of the underlying device. It is not neccesarily
// the same as the one during NewDevice. However, it is guaranteed to
// point at the same device (symlinks may be resolved).
func (dev *Device) GetSyspath() string {
	cstr := C.xwii_iface_get_syspath(dev.cptr)
	if cstr == nil {
		return ""
	}
	return C.GoString(cstr)
}

// FD returns the file-descriptor to notify readiness. If multiple file-descriptors
// are used internally, they are multi-plexed through an epoll descriptor.
// Therefore, this always returns the same single file-descriptor. You need to
// watch this for readable-events (POLLIN/EPOLLIN) and call
// Poll() whenever it is readable.
func (dev *Device) FD() int {
	return int(C.xwii_iface_get_fd(dev.cptr))
}

// Watch sets whether hotplug events should be reported or not. By default, no
// hotplug events are reported so this is off.
//
// Note that this requires a separate udev-monitor for each device. Therefore,
// if your application uses its own udev-monitor, you should instead integrate
// the hotplug-detection into your udev-monitor.
func (dev *Device) Watch(hotplug bool) error {
	ret := C.xwii_iface_watch(dev.cptr, C.bool(hotplug))
	return cError(ret)
}

// Open all the requested interfaces. If InterfaceWritable is also set,
// the interfaces are opened with write-access. Note that interfaces that are
// already opened are ignored and not touched.
// If any interface fails to open, this function still tries to open the other
// requested interfaces and then returns the error afterwards. Hence, if this
// function fails, you should use Opened() to get a bitmask of opened
// interfaces and see which failed (if that is of interest).
//
// Note that interfaces may be closed automatically during runtime if the
// kernel removes the interface or on error conditions. You always get an
// EventWatch event which you should react on. This is returned
// regardless whether Watch() was enabled or not.
func (dev *Device) Open(ifaces InterfaceType) error {
	ret := C.xwii_iface_open(dev.cptr, C.uint(ifaces))
	return cError(ret)
}

// Close interfaces on this device.
func (dev *Device) Close(ifaces InterfaceType) {
	C.xwii_iface_close(dev.cptr, C.uint(ifaces))
}

// Opened returns a bitmask of opened interfaces. Interfaces may be closed due to
// error-conditions at any time. However, interfaces are never opened
// automatically.
//
// You will get notified whenever this bitmask changes, except on explicit
// calls to Open() and Close(). See the EventWatch event for more information.
func (dev *Device) Opened() InterfaceType {
	ret := C.xwii_iface_opened(dev.cptr)
	return InterfaceType(ret)
}

// Available returns a bitmask of available devices. These devices can be opened and are
// guaranteed to be present on the hardware at this time. If you watch your
// device for hotplug events you will get notified whenever this bitmask changes.
// See the WatchEvent event for more information.
func (dev *Device) Available() InterfaceType {
	ret := C.xwii_iface_available(dev.cptr)
	return InterfaceType(ret)
}

// Poll for incoming events.
//
// You should call this whenever the file-descriptor returned by
// FD is reported as being readable. This function will perform
// all non-blocking outstanding tasks and then return.
//
// This function always performs any background tasks and outgoing event-writes
// if they don't block. It returns an error if they fail. This function then tries to
// read a single incoming event. If no event is available, it returns no error but sets continue-flag low
// and you should watch the file-desciptor again until it is readable. Otherwise, you should call this
// function in a row as long as it returns 0.
//
// It returns the event or nil if an error occured, the continue-flag whether a new event can be polled right away and
// optionally and error, if the error is ErrRetry, consider polling again for new events.
func (dev *Device) Poll() (Event, bool, error) {
	var ev C.struct_xwii_event
	ret := C.xwii_iface_dispatch(dev.cptr, &ev, C.size_t(unsafe.Sizeof(ev)))
	if ret == -C.EAGAIN {
		return nil, false, ErrPollAgain
	}
	if ret != 0 {
		return nil, false, cError(ret)
	}
	event := makeEvent(ev)
	if event == nil {
		return nil, false, os.ErrInvalid
	}
	return event, true, nil
}

// Work makes some background work.
//
// You should call this whenever the file-descriptor returned by
// FD() is reported as being readable. This function will perform
// all non-blocking outstanding tasks and then return.
//
// This function always performs any background tasks and outgoing event-writes
// if they don't block. It returns an error if they fail.
func (dev *Device) Work() error {
	ret := C.xwii_iface_dispatch(dev.cptr, nil, 0)
	return cError(ret)
}

// Rumble sets the rumble motor.
//
// This requires the core-interface to be opened in writable mode.
func (dev *Device) Rumble(state bool) error {
	ret := C.xwii_iface_watch(dev.cptr, C.bool(state))
	return cError(ret)
}

// GetLED reads the LED state for the given LED.
//
// LEDs are a static interface that does not have to be opened first.
func (dev *Device) GetLED() (Led, error) {
	var (
		state  C.bool
		ret    C.int
		result Led
	)
	for i := range 4 {
		ret = C.xwii_iface_get_led(dev.cptr, C.uint(i), &state)
		if ret != 0 {
			return 0, cError(ret)
		}
		if state {
			result |= 1 << i
		}
	}
	return result, nil
}

// SetLED writes the LED state for the given LED.
//
// LEDs are a static interface that does not have to be opened first.
func (dev *Device) SetLED(leds Led) error {
	var (
		ret C.int
	)
	for i := range 4 {
		ret = C.xwii_iface_set_led(dev.cptr, C.uint(i), leds&(1<<i) != 0)
		if ret != 0 {
			return cError(ret)
		}
	}
	return nil
}

// GetBattery reads the current battery capacity. The capacity is represented as percentage, thus the return value is an integer between 0 and 100.
//
// Batteries are a static interface that does not have to be opened first.
func (dev *Device) GetBattery() (uint8, error) {
	var state C.uint8_t
	ret := C.xwii_iface_get_battery(dev.cptr, &state)
	return uint8(state), cError(ret)
}

// GetDevType returns the device type. If the device type cannot be determined,
// it returns "unknown" and the corresponding error.
//
// This is a static interface that does not have to be opened first.
func (dev *Device) GetDevType() (string, error) {
	var devtype *C.char
	ret := C.xwii_iface_get_devtype(dev.cptr, &devtype)
	if ret != 0 {
		return "<unknown>", cError(ret)
	}
	return cStringCopy(devtype), nil
}

// GetExtension returns the extension type. If no extension is connected or the
// extension cannot be determined, it returns a string "none" and the corresponding error.
//
// This is a static interface that does not have to be opened first.
func (dev *Device) GetExtension() (string, error) {
	var ext *C.char
	ret := C.xwii_iface_get_extension(dev.cptr, &ext)
	if ret != 0 {
		return "none", cError(ret)
	}
	return cStringCopy(ext), nil
}

// SetMPNormalization sets Motion-Plus normalization and calibration values. The Motion-Plus sensor is very
// sensitive and may return really crappy values. This interfaces allows to
// apply 3 absolute offsets x, y and z which are subtracted from any MP data
// before it is returned to the application. That is, if you set these values
// to 0, this has no effect (which is also the initial state).
//
// The calibration factor is used to perform runtime calibration. If
// it is 0 (the initial state), no runtime calibration is performed. Otherwise,
// the factor is used to re-calibrate the zero-point of MP data depending on MP
// input. This is an angoing calibration which modifies the internal state of
// the x, y and z values.
func (dev *Device) SetMPNormalization(x, y, z, factor int32) {
	C.xwii_iface_set_mp_normalization(dev.cptr, C.int32_t(x), C.int32_t(y), C.int32_t(z), C.int32_t(factor))
}

// GetMPNormalization reads the Motion-Plus normalization and calibration values. Please see
// SetMPNormalization() how this is handled.
//
// Note that if the calibration factor is not 0, the normalization values may
// change depending on incoming MP data. Therefore, the data read via this
// function may differ from the values that you wrote to previously. However,
// apart from applied calibration, these value are the same as were set
// previously via SetMPNormalization() and you can feed them back
// in later.
func (dev *Device) GetMPNormalization() (x, y, z, factor int32) {
	C.xwii_iface_get_mp_normalization(dev.cptr, (*C.int32_t)(&x), (*C.int32_t)(&y), (*C.int32_t)(&z), (*C.int32_t)(&factor))
	return
}

func (dev *Device) String() string {
	var w strings.Builder
	w.WriteString("xwiimote-device ")
	devtype, _ := dev.GetDevType()
	w.WriteString(devtype)
	ext, _ := dev.GetExtension()
	if ext != "none" && ext != "" {
		w.WriteString(" with ")
		w.WriteString(ext)
	}
	w.WriteString(" at ")
	w.WriteString(dev.GetSyspath())
	return w.String()
}
