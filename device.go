// Package xwiimote has bindings for libxwiimote, a library to read and control inputs on a Nintendo WiiMote and accessories.
package xwiimote

// #include <linux/input.h>
// #include <errno.h>
import "C"
import (
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/friedelschoen/go-xwiimote/pkg/udev"
)

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

	//  epoll file descriptor
	efd int
	//  main udev device
	dev *udev.Device
	//  udev monitor
	umon *udev.Monitor

	// open interfaces -- node -> interface
	ifs map[string]Interface
	// available interfaces -- node -> name
	availIfs map[string]string
	//  device type attribute
	devtypeAttr string
	//  extension attribute
	extensionAttr string
	//  battery capacity attribute
	batteryAttr string
	//  led brightness attributes
	ledAttrs [4]string

	//  motion plus normalization
	mpNormalizer     Vec3 // event_abs
	mpNormaizeFactor int32
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
	dev := udev.NewDeviceFromSyspath(syspath)
	if dev == nil {
		return nil, os.ErrInvalid
	}

	return newDeviceFromUdev(dev)
}

func newDeviceFromUdev(dev *udev.Device) (*Device, error) {
	var d Device
	d.poller = newPoller(&d)
	d.dev = dev

	driver := d.dev.Driver()
	subs := d.dev.Subsystem()
	if driver != "wiimote" || subs != "hid" {
		return nil, os.ErrInvalid
	}
	syspath := dev.Syspath()
	d.devtypeAttr = path.Join(syspath, "devtype")
	d.extensionAttr = path.Join(syspath, "extension")

	var err error
	d.efd, err = syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	if err := d.readNodes(); err != nil {
		syscall.Close(d.efd)
		return nil, err
	}

	runtime.AddCleanup(&d, func(fd int) { syscall.Close(d.efd) }, d.efd)

	return &d, nil
}

// Scan the device \dev for child input devices and update our device-node
// cache with the new information. This is called during device setup to
// find all /dev/input/eventX nodes for all currently available interfaces.
// We also cache attribute paths for sub-devices like LEDs or batteries.
//
// When called during hotplug-events, this updates all currently known
// information and removes nodes that are no longer present.
func (dev *Device) readNodes() error {
	e := udev.NewEnumerate()

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

	if dev.availIfs == nil {
		dev.availIfs = make(map[string]string)
	}
	if dev.ifs == nil {
		dev.ifs = make(map[string]Interface)
	}

	// The returned list is sorted. So we first get an inputXY entry,
	// possibly followed by the inputXY/eventXY entry. We remember the type
	// of a found inputXY entry, and check the next list-entry, whether
	// it's an eventXY entry. If it is, we save the node, otherwise, it's
	// skipped.
	// For other subsystems we simply cache the attribute paths.
	var prevIf string
	matches, err := e.Devices()
	if err != nil {
		return err
	}
	for syspath := range matches {
		d := udev.NewDeviceFromSyspath(syspath)

		name := d.Sysname()
		switch d.Subsystem() {
		case "input":
			if strings.HasPrefix(name, "input") {
				name := d.SysattrValue("name")
				if name == "" {
					continue
				}
				prevIf = name
			} else if strings.HasPrefix(name, "event") {
				if prevIf == "" {
					continue
				}
				node := d.Devnode()
				if node == "" {
					continue
				}
				dev.availIfs[prevIf] = node
			}
		case "leds":
			num := syspath[len(syspath)-1]
			if num < '0' || num > '3' {
				continue
			}

			if dev.ledAttrs[num-'0'] != "" {
				continue
			}
			dev.ledAttrs[num-'0'] = path.Join(syspath, "brightness")
		case "power_supply":
			if dev.batteryAttr != "" {
				continue
			}
			dev.batteryAttr = path.Join(syspath, "capacity")
		}
	}

	// close no longer available ifaces
	for _, iff := range dev.ifs {
		if _, ok := dev.availIfs[iff.Name()]; !ok {
			dev.CloseInterfaces(iff)
		}
	}

	return nil
}

// CloseInterfaces closes one or more interfaces on this device.
func (dev *Device) CloseInterfaces(ifaces ...Interface) error {
	if len(ifaces) == 0 {
		return nil
	}
	var errs []error
	for _, iface := range ifaces {
		if err := iface.close(); err != nil {
			errs = append(errs, err)
			continue
		}
		delete(dev.ifs, iface.Node())
	}
	return dev.readNodes()
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
	if !hotplug {
		//  remove device watch descriptor

		if dev.umon == nil {
			return nil
		}

		fd := dev.umon.GetFD()
		syscall.EpollCtl(dev.efd, syscall.EPOLL_CTL_DEL, fd, nil)
		dev.umon = nil
		return nil
	}

	//  add device watch descriptor
	if dev.umon != nil {
		return nil
	}

	dev.umon = udev.NewMonitorFromNetlink("udev")
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
	syscall.SetNonblock(fd, true)

	var ep syscall.EpollEvent
	ep.Events = syscall.EPOLLIN
	ep.Fd = int32(fd)

	if err := syscall.EpollCtl(dev.efd, syscall.EPOLL_CTL_ADD, fd, &ep); err != nil {
		return err
	}

	return nil
}

// OpenInterfaces all the requested interfaces. If InterfaceWritable is also set,
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
func (dev *Device) OpenInterfaces(wr bool, ifaces ...Interface) error {
	for _, iface := range ifaces {
		node, ok := dev.availIfs[iface.Name()]
		if !ok {
			continue
		}
		if err := iface.open(dev, node, wr); err != nil {
			return err
		}
		dev.ifs[node] = iface
	}
	return nil
}

// IsAvailable returns a bitmask of available devices. These devices can be opened and are
// guaranteed to be present on the hardware at this time. If you watch your
// device for hotplug events you will get notified whenever this bitmask changes.
// See the WatchEvent event for more information.
func (dev *Device) Available(iface Interface) bool {
	_, ok := dev.availIfs[iface.Name()]
	return ok
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

	//  write outgoing events here
	n, err := syscall.EpollWait(dev.efd, ep[:], 0)
	if err != nil {
		return nil, false, err
	}
	for _, pollev := range ep[:n] {
		ev, err := dev.dispatchEvent(pollev.Fd, pollev.Events)
		if err != nil {
			return nil, false, err
		}
		if ev != nil {
			return ev, true, nil
		}
	}

	return nil, false, ErrPollAgain
}

// GetLED reads the LED state for the given LED.
//
// LEDs are a static interface that does not have to be opened first.
func (dev *Device) GetLED() (result Led, _ error) {
	for i := range 4 {
		cont, err := os.ReadFile(dev.ledAttrs[i])
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
		if err := os.WriteFile(dev.ledAttrs[i], []byte(cont), 0); err != nil {
			return err
		}
	}
	return nil
}

// GetBattery reads the current battery capacity. The capacity is represented as percentage, thus the return value is an integer between 0 and 100.
//
// Batteries are a static interface that does not have to be opened first.
func (dev *Device) GetBattery() (uint, error) {
	cont, err := os.ReadFile(dev.batteryAttr)
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
	cont, err := os.ReadFile(dev.devtypeAttr)
	return strings.TrimSpace(string(cont)), err
}

// GetExtension returns the extension type. If no extension is connected or the
// extension cannot be determined, it returns a string "none" and the corresponding error.
//
// This is a static interface that does not have to be opened first.
func (dev *Device) GetExtension() (string, error) {
	cont, err := os.ReadFile(dev.extensionAttr)
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
	dev.mpNormalizer.X = x * 100
	dev.mpNormalizer.Y = y * 100
	dev.mpNormalizer.Z = z * 100
	dev.mpNormaizeFactor = factor
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
	return dev.mpNormalizer.X / 100,
		dev.mpNormalizer.Y / 100,
		dev.mpNormalizer.Z / 100,
		dev.mpNormaizeFactor
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
