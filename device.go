// Package xwiimote has bindings for libxwiimote, a library to read and control inputs on a Nintendo WiiMote and accessories.
package xwiimote

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
// #include <linux/input.h>
// #include <errno.h>
//
// unsigned int eviocgname(size_t sz) { return EVIOCGNAME(sz); }
import "C"
import (
	"errors"
	"io"
	"iter"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/friedelschoen/go-xwiimote/pkg/udev"
)

var (
	ErrInvalidDevice = errors.New("device is not a wiimote")
)

// InterfaceType describes a single interface. These are bit-masks that can be
// binary-ORed. If an interface does not provide such a constant, it is static
// and can be used without opening/closing it.
type InterfaceType int

const (
	// Core interface
	InterfaceCore InterfaceType = 0x000001
	// Accelerometer interface
	InterfaceAccel InterfaceType = 0x000002
	// IR interface
	InterfaceIR InterfaceType = 0x000004
	// MotionPlus extension interface
	InterfaceMotionPlus InterfaceType = 0x000100
	// Nunchuk extension interface
	InterfaceNunchuk InterfaceType = 0x000200
	// ClassicController extension interface
	InterfaceClassicController InterfaceType = 0x000400
	// BalanceBoard extension interface
	InterfaceBalanceBoard InterfaceType = 0x000800
	// ProController extension interface
	InterfaceProController InterfaceType = 0x001000
	// Drums extension interface
	InterfaceDrums InterfaceType = 0x002000
	// Guitar extension interface
	InterfaceGuitar InterfaceType = 0x004000
	// Special flag ORed with all valid interfaces
	InterfaceAll InterfaceType = InterfaceCore |
		InterfaceAccel |
		InterfaceIR |
		InterfaceMotionPlus |
		InterfaceNunchuk |
		InterfaceClassicController |
		InterfaceBalanceBoard |
		InterfaceProController |
		InterfaceDrums |
		InterfaceGuitar
	// Special flag which causes the interfaces to be opened writable
	InterfaceWritable InterfaceType = 0x010000
)

// Name returns the original name of that interface.
func (dev InterfaceType) Name() string {
	switch dev {
	case InterfaceCore:
		return "Nintendo Wii Remote"
	case InterfaceAccel:
		return "Nintendo Wii Remote Accelerometer"
	case InterfaceIR:
		return "Nintendo Wii Remote IR"
	case InterfaceMotionPlus:
		return "Nintendo Wii Remote Motion Plus"
	case InterfaceNunchuk:
		return "Nintendo Wii Remote Nunchuk"
	case InterfaceClassicController:
		return "Nintendo Wii Remote Classic Controller"
	case InterfaceBalanceBoard:
		return "Nintendo Wii Remote Balance Board"
	case InterfaceProController:
		return "Nintendo Wii Remote Pro Controller"
	case InterfaceDrums:
		return "Nintendo Wii Remote Drums"
	case InterfaceGuitar:
		return "Nintendo Wii Remote Guitar"
	default:
		return "Nintendo Wii Remote Unknown?"
	}
}

// Name returns the original name of that interface.
func typeFromName(name string) InterfaceType {
	switch name {
	case "Nintendo Wii Remote":
		return InterfaceCore
	case "Nintendo Wii Remote Accelerometer":
		return InterfaceAccel
	case "Nintendo Wii Remote IR":
		return InterfaceIR
	case "Nintendo Wii Remote Motion Plus":
		return InterfaceMotionPlus
	case "Nintendo Wii Remote Nunchuk":
		return InterfaceNunchuk
	case "Nintendo Wii Remote Classic Controller":
		return InterfaceClassicController
	case "Nintendo Wii Remote Balance Board":
		return InterfaceBalanceBoard
	case "Nintendo Wii Remote Pro Controller":
		return InterfaceProController
	case "Nintendo Wii Remote Drums":
		return InterfaceDrums
	case "Nintendo Wii Remote Guitar":
		return InterfaceGuitar
	default:
		return 0
	}
}

// Led described a Led of an device. The leds are counted left-to-right and can be OR'ed together.
type Led uint

const (
	Led1 Led = 1 << iota
	Led2
	Led3
	Led4
)

type Interface struct {
	// Type of interface, may not be bit-or'ed
	Type InterfaceType
	// Device node as /dev/input/eventX or ""
	node string
	// Open file or nil
	fd *os.File
	// Temporary state during device detection
	available bool
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func iterIfaceTypes[T Integer](ifaces T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := T(1); i <= ifaces; i >>= 1 {
			if ifaces&i != 0 && !yield(i) {
				return
			}
		}
	}
}

// Device describes the communication with a single device. That is, you
// create one for each device you use. All sub-interfaces are opened on this
// object.
type Device struct {
	poller[Event]
	// cptr *C.struct_xwii_iface

	/* epoll file descriptor */
	efd int
	/* udev context */
	udev udev.Udev
	/* main udev device */
	dev *udev.Device
	/* udev monitor */
	umon *udev.Monitor

	/* bitmask of open interfaces */
	// ifaces uint ; . ifs
	/* interfaces */
	ifs map[InterfaceType]*Interface
	/* device type attribute */
	devtype_attr string
	/* extension attribute */
	extension_attr string
	/* battery capacity attribute */
	battery_attr string
	/* led brightness attributes */
	led_attrs [4]string

	/* rumble-id for base-core interface force-feedback or -1 */
	rumble_id int
	rumble_fd *os.File
	/* accelerometer data cache */
	accel_cache EventAccel
	/* IR data cache */
	ir_cache EventIR
	/* balance board weight cache */
	bboard_cache EventBalanceBoard
	/* motion plus cache */
	mp_cache EventMotionPlus
	/* motion plus normalization */
	mp_normalizer       Vec3 // event_abs
	mp_normalize_factor int32
	/* pro controller cache */
	pro_cache EventProControllerMove
	/* classic controller cache */
	classic_cache EventClassicControllerMove
	/* nunchuk cache */
	nunchuk_cache EventNunchukMove
	/* drums cache */
	drums_cache EventDrumsMove
	/* guitar cache */
	guitar_cache EventGuitarMove
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
	d := new(Device)
	d.poller = newPoller(d)

	// const char *driver, *subs;
	// int ret, i;

	if syspath == "" {
		return nil, os.ErrInvalid
	}

	d.rumble_id = -1

	var err error
	d.efd, err = syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}

	d.dev = d.udev.NewDeviceFromSyspath(syspath)

	driver := d.dev.Driver()
	subs := d.dev.Subsystem()
	if driver != "wiimote" || subs != "hid" {
		syscall.Close(d.efd)
		return nil, ErrInvalidDevice
	}
	d.devtype_attr = path.Join(syspath, "devtype")
	d.extension_attr = path.Join(syspath, "extension")

	if err := d.readNodes(); err != nil {
		syscall.Close(d.efd)
		return nil, err
	}

	return d, nil
}

/*
 * Scan the device \dev for child input devices and update our device-node
 * cache with the new information. This is called during device setup to
 * find all /dev/input/eventX nodes for all currently available interfaces.
 * We also cache attribute paths for sub-devices like LEDs or batteries.
 *
 * When called during hotplug-events, this updates all currently known
 * information and removes nodes that are no longer present.
 */
func (dev *Device) readNodes() error {
	e := dev.udev.NewEnumerate()

	if err := e.AddMatchSubsystem("input"); err != nil {
		return err
	}
	if err := e.AddMatchSubsystem("leds"); err != nil {
		return err
	}
	if err := e.AddMatchSubsystem("power_supply"); err != nil {
		return err
	}
	if err := e.AddMatchParent(dev.dev); err != nil {
		return err
	}

	for _, i := range dev.ifs {
		i.available = false
	}

	/* The returned list is sorted. So we first get an inputXY entry,
	 * possibly followed by the inputXY/eventXY entry. We remember the type
	 * of a found inputXY entry, and check the next list-entry, whether
	 * it's an eventXY entry. If it is, we save the node, otherwise, it's
	 * skipped.
	 * For other subsystems we simply cache the attribute paths. */
	prev_if := InterfaceType(0)
	matches, err := e.Devices()
	if err != nil {
		return err
	}
	for name := range matches {
		d := dev.udev.NewDeviceFromSyspath(name)
		tif := prev_if
		prev_if = 0

		syspath := d.Syspath()
		switch d.Subsystem() {
		case "input":
			if strings.HasPrefix(name, "input") {
				name := d.SysattrValue("name")
				if name == "" {
					continue
				}
				tif = typeFromName(name)
				if tif > 0 {
					prev_if = tif
				}
			} else if strings.HasPrefix(name, "event") {
				if tif == 0 {
					continue
				}
				node := d.Devnode()
				if node == "" {
					continue
				}
				if iff, ok := dev.ifs[tif]; ok {
					if iff.node == node {
						iff.available = true
					} else {
						delete(dev.ifs, tif)
					}
				} else {
					dev.ifs[tif] = &Interface{
						node:      node,
						available: true,
					}
				}
			}
		case "leds":
			num := syspath[len(syspath)-1]
			if num < '0' || num > '3' {
				continue
			}

			if dev.led_attrs[num-'0'] != "" {
				continue
			}
			dev.led_attrs[num-'0'] = path.Join(syspath, "brightness")
		case "power_supply":
			if dev.battery_attr != "" {
				continue
			}
			dev.battery_attr = path.Join(syspath, "capacity")
		}
	}

	/* close no longer available ifaces */
	ifs := InterfaceType(0)
	for iname, iff := range dev.ifs {
		if !iff.available {
			ifs |= iname
		}
	}
	dev.closeInterface(ifs)

	return nil
}

// Close interfaces on this device.
func (dev *Device) CloseInterface(ifaces InterfaceType) {
	ifaces &= InterfaceAll
	if ifaces == 0 {
		return
	}

	for iface := range iterIfaceTypes(ifaces) {
		dev.closeInterface(iface)
	}

	for iface := range iterIfaceTypes(ifaces & (InterfaceCore | InterfaceProController)) {
		if iff, ok := dev.ifs[iface]; ok && iff.fd == dev.rumble_fd {
			dev.rumble_id = -1
			dev.rumble_fd = nil
			break
		}
	}
}

func (dev *Device) closeInterface(tif InterfaceType) {
	iff, ok := dev.ifs[tif]
	if !ok {
		return
	}
	if iff.fd == nil {
		return
	}

	syscall.EpollCtl(dev.efd, syscall.EPOLL_CTL_DEL, int(iff.fd.Fd()), nil)
	iff.fd.Close()
	delete(dev.ifs, tif)
}

// Free unreferences the devices and frees the underlying structure.
// Calling Free is not mandatory and is done automatically by default.
func (dev *Device) Close() error {
	return syscall.Close(dev.efd)
}

// GetSyspath returns the sysfs path of the underlying device. It is not neccesarily
// the same as the one during NewDevice. However, it is guaranteed to
// point at the same device (symlinks may be resolved).
func (dev *Device) GetSyspath() string {
	return dev.dev.Syspath()
}

// FD returns the file-descriptor to notify readiness. If multiple file-descriptors
// are used internally, they are multi-plexed through an epoll descriptor.
// Therefore, this always returns the same single file-descriptor. You need to
// watch this for readable-events (POLLIN/EPOLLIN) and call
// Poll() whenever it is readable.
func (dev *Device) FD() int {
	return dev.efd
}

// Watch sets whether hotplug events should be reported or not. By default, no
// hotplug events are reported so this is off.
//
// Note that this requires a separate udev-monitor for each device. Therefore,
// if your application uses its own udev-monitor, you should instead integrate
// the hotplug-detection into your udev-monitor.
func (dev *Device) Watch(hotplug bool) error {
	// int fd, ret, set;
	// struct epoll_event ep;

	if !hotplug {
		/* remove device watch descriptor */

		if dev.umon == nil {
			return nil
		}

		fd := dev.umon.GetFD()
		syscall.EpollCtl(dev.efd, syscall.EPOLL_CTL_DEL, fd, nil)
		dev.umon = nil
		return nil
	}

	/* add device watch descriptor */
	if dev.umon != nil {
		return nil
	}

	dev.umon = dev.udev.NewMonitorFromNetlink("udev")
	if err := dev.umon.FilterAddMatchSubsystem("input"); err != nil {
		return err
	}
	if err := dev.umon.FilterAddMatchSubsystem("hid"); err != nil {
		return err
	}
	if err := dev.umon.EnableReceiving(); err != nil {
		return err
	}

	fd := dev.umon.GetFD()
	syscall.SetNonblock(int(fd), true)

	var ep syscall.EpollEvent
	ep.Events = syscall.EPOLLIN
	ep.Fd = int32(dev.efd)

	if err := syscall.EpollCtl(dev.efd, syscall.EPOLL_CTL_ADD, fd, &ep); err != nil {
		return err
	}

	return nil
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
	wr := ifaces&InterfaceWritable > 0
	ifaces &= InterfaceAll
	for name := range dev.ifs {
		ifaces &= ^name
	}
	if ifaces == 0 {
		return nil
	}

	for iface := range iterIfaceTypes(ifaces) {
		if err := dev.openOneInterface(iface, wr); err != nil {
			return err
		}
	}

	for iface := range iterIfaceTypes(ifaces & (InterfaceCore | InterfaceProController)) {
		if err := dev.uploadRumble(dev.ifs[iface].fd); err != nil {
			return err
		}
	}

	return nil
}

func (dev *Device) openOneInterface(tif InterfaceType, wr bool) error {
	// char name[256];
	// struct epoll_event ep;
	// unsigned int flags;
	// int fd, err;

	iff, ok := dev.ifs[tif]
	if !ok {
		return os.ErrNotExist
	}

	if iff.fd != nil {
		return nil
	}

	flags := syscall.O_NONBLOCK | syscall.O_CLOEXEC
	if wr {
		flags |= os.O_RDWR
	}
	fd, err := os.OpenFile(iff.node, flags, 0)
	if err != nil {
		return err
	}

	// if ioctl(fd, EVIOCGNAME(sizeof(name)), name) < 0 {
	// 	close(fd)
	// 	return -errno
	// }

	// name[sizeof(name)-1] = 0
	// if strcmp(if_to_name_table[tif], name) {
	// 	close(fd)
	// 	return -ENODEV
	// }

	var name [256]byte
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), uintptr(C.eviocgname(C.size_t(len(name)))), uintptr(unsafe.Pointer(&name[0]))); err != 0 {
		return err
	}

	var ep syscall.EpollEvent
	ep.Events = syscall.EPOLLIN
	ep.Fd = int32(fd.Fd())
	if err := syscall.EpollCtl(dev.efd, syscall.EPOLL_CTL_ADD, int(fd.Fd()), &ep); err != nil {
		fd.Close()
		return err
	}

	iff.fd = fd
	return nil
}

/*
 * Upload the generic rumble event to the device. This may later be used for
 * force-feedback effects. The event id is safed for later use.
 */
func (dev *Device) uploadRumble(fd *os.File) error {
	effect := C.struct_ff_effect{
		_type: C.FF_RUMBLE,
		id:    -1,
	}

	rmb := (*C.struct_ff_rumble_effect)(unsafe.Pointer(&effect.u))
	rmb.strong_magnitude = 1

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), C.EVIOCSFF, uintptr(unsafe.Pointer(&effect))); err != 0 {
		return err
	}
	dev.rumble_id = int(effect.id)
	dev.rumble_fd = fd
	return nil
}

// Opened returns a bitmask of opened interfaces. Interfaces may be closed due to
// error-conditions at any time. However, interfaces are never opened
// automatically.
//
// You will get notified whenever this bitmask changes, except on explicit
// calls to Open() and Close(). See the EventWatch event for more information.
func (dev *Device) Opened() InterfaceType {
	var ifaces InterfaceType
	for name, iff := range dev.ifs {
		if iff.fd != nil {
			ifaces |= name
		}
	}
	return ifaces
}

// Available returns a bitmask of available devices. These devices can be opened and are
// guaranteed to be present on the hardware at this time. If you watch your
// device for hotplug events you will get notified whenever this bitmask changes.
// See the WatchEvent event for more information.
func (dev *Device) Available() InterfaceType {
	var ifaces InterfaceType
	for name := range dev.ifs {
		ifaces |= name
	}
	return ifaces
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
	var ep [32]syscall.EpollEvent
	// int ret, i;
	// size_t siz;
	// struct xwii_event ev;

	/* write outgoing events here */

	n, err := syscall.EpollWait(dev.efd, ep[:], 0)
	if err != nil {
		return nil, false, err
	}

	for i := 0; i < n; i++ {
		ev, err := dev.dispatchEvent(ep[i].Fd)
		if err != nil {
			return nil, false, err
		}
		if ev != nil {
			return ev, true, nil
		}
	}

	return nil, false, ErrPollAgain
}

// Rumble sets the rumble motor.
//
// This requires the core-interface to be opened in writable mode.
func (dev *Device) Rumble(state bool) error {

	if dev.rumble_fd == nil || dev.rumble_id < 0 {
		return os.ErrInvalid
	}

	var ev C.struct_input_event
	ev._type = C.EV_FF
	ev.code = C.ushort(dev.rumble_id)
	if state {
		ev.value = 1
	}

	buf := unsafe.Slice((*byte)(unsafe.Pointer(&ev)), unsafe.Sizeof(ev))

	n, err := dev.rumble_fd.Write(buf)
	if err != nil {
		return err
	}
	if n != int(unsafe.Sizeof(ev)) {
		return io.ErrShortWrite
	}
	return nil

}

// GetLED reads the LED state for the given LED.
//
// LEDs are a static interface that does not have to be opened first.
func (dev *Device) GetLED() (result Led, _ error) {
	for i := range 4 {
		cont, err := os.ReadFile(dev.led_attrs[i])
		if err != nil {
			return 0, err
		}
		if strings.TrimSpace(string(cont)) == "1" {
			result |= 1 << i
		}
	}
	return result, nil
}

// SetLED writes the LED state for the given LED.
//
// LEDs are a static interface that does not have to be opened first.
func (dev *Device) SetLED(leds Led) error {
	for i := range 4 {
		state := leds&(1<<i) != 0

		cont := "0\n"
		if state {
			cont = "1\n"
		}
		if err := os.WriteFile(dev.led_attrs[i], []byte(cont), 0); err != nil {
			return err
		}
	}
	return nil
}

// GetBattery reads the current battery capacity. The capacity is represented as percentage, thus the return value is an integer between 0 and 100.
//
// Batteries are a static interface that does not have to be opened first.
func (dev *Device) GetBattery() (uint, error) {
	cont, err := os.ReadFile(dev.battery_attr)
	if err != nil {
		return 0, nil
	}

	cap, err := strconv.Atoi(strings.TrimSpace(string(cont)))
	return uint(cap), err
}

// GetDevType returns the device type. If the device type cannot be determined,
// it returns "unknown" and the corresponding error.
//
// This is a static interface that does not have to be opened first.
func (dev *Device) GetDevType() (string, error) {
	cont, err := os.ReadFile(dev.devtype_attr)
	return strings.TrimSpace(string(cont)), err
}

// GetExtension returns the extension type. If no extension is connected or the
// extension cannot be determined, it returns a string "none" and the corresponding error.
//
// This is a static interface that does not have to be opened first.
func (dev *Device) GetExtension() (string, error) {
	cont, err := os.ReadFile(dev.extension_attr)
	return strings.TrimSpace(string(cont)), err
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
	dev.mp_normalizer.X = x * 100
	dev.mp_normalizer.Y = y * 100
	dev.mp_normalizer.Z = z * 100
	dev.mp_normalize_factor = factor
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
	return dev.mp_normalizer.X / 100,
		dev.mp_normalizer.Y / 100,
		dev.mp_normalizer.Z / 100,
		dev.mp_normalize_factor
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
