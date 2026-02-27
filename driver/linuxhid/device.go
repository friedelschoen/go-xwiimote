// Package wiimote has bindings for libwiimote, a library to read and control inputs on a Nintendo WiiMote and accessories.
package linuxhid

// #include <linux/input.h>
// #include <errno.h>
import "C"
import (
	"errors"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/discovery"
	"github.com/friedelschoen/go-wiimote/internal/common"
)

const debugfs = "/sys/kernel/debug"

// Device describes the communication with a single device. That is, you
// create one for each device you use. All sub-interfaces are opened on this
// object.
type Device struct {
	wiimote.Poller[wiimote.Event]

	//  epoll file descriptor
	efd int
	//  main udev device
	dev wiimote.DeviceInfo
	//  udev monitor
	umon wiimote.DeviceMonitor

	// open interfaces -- node -> interface
	ifs map[string]Interface
	// available interfaces -- node -> name
	availIfs map[wiimote.InterfaceKind]string
	// device type attribute
	devtypeAttr string
	// extension attribute
	extensionAttr string
	// battery capacity attribute
	batteryAttr string
	// led brightness attributes
	ledAttrs [4]string
	// buffers internal events
	moreEvents chan wiimote.Event
}

// NewDevice creates a new device object. No interfaces on the device are opened by
// default.
//
// syspath must be a valid path to a wiimote device, either
// retrieved via a Monitor, an Enumerator or via udev directly. It must point to
// the hid device, which is normally /sys/bus/hid/devices/[dev].
//
// The object and underlying structure is freed automatically by default.
func NewDevice(dev wiimote.DeviceInfo) (*Device, error) {
	var d Device
	d.Poller = common.NewPoller(&d)
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
	d.moreEvents = make(chan wiimote.Event, 128)

	d.umon = discovery.NewMonitor()
	if err := d.umon.FilterAddMatchSubsystem("input"); err != nil {
		syscall.Close(d.efd)
		return nil, err
	}
	if err := d.umon.FilterAddMatchSubsystem("hid"); err != nil {
		syscall.Close(d.efd)
		return nil, err
	}
	if err := d.umon.EnableReceiving(); err != nil {
		syscall.Close(d.efd)
		return nil, err
	}

	fd := d.umon.FD()
	syscall.SetNonblock(fd, true)

	var ep syscall.EpollEvent
	ep.Events = syscall.EPOLLIN
	ep.Fd = int32(fd)

	if err := syscall.EpollCtl(d.efd, syscall.EPOLL_CTL_ADD, fd, &ep); err != nil {
		syscall.Close(d.efd)
		return nil, err
	}

	runtime.AddCleanup(&d, func(fd int) { syscall.Close(fd) }, d.efd)

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
	e := discovery.NewEnumerate()

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
		dev.availIfs = make(map[wiimote.InterfaceKind]string)
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
	for d := range matches {
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
				kind, ok := InterfaceKindFromName(prevIf)
				if ok {
					dev.availIfs[kind] = node
				}
			}
		case "leds":
			num := d.Syspath()[len(d.Syspath())-1]
			if num < '0' || num > '3' {
				continue
			}

			if dev.ledAttrs[num-'0'] != "" {
				continue
			}
			dev.ledAttrs[num-'0'] = path.Join(d.Syspath(), "brightness")
		case "power_supply":
			if dev.batteryAttr != "" {
				continue
			}
			dev.batteryAttr = path.Join(d.Syspath(), "capacity")
		}
	}

	// close no longer available ifaces
	for _, iff := range dev.ifs {
		if _, ok := dev.availIfs[iff.Kind()]; !ok {
			iff.Close()
		}
	}

	return nil
}

// FD returns the file-descriptor to notify readiness. If multiple file-descriptors
// are used internally, they are multi-plexed through an epoll descriptor.
// Therefore, this always returns the same single file-descriptor. You need to
// watch this for readable-events (POLLIN/EPOLLIN) and call
// Poll() whenever it is readable.
func (dev *Device) FD() int {
	return dev.efd
}

// Syspath returns the sysfs path of the underlying device. It is not neccesarily
// the same as the one during NewDevice. However, it is guaranteed to
// point at the same device (symlinks may be resolved).
func (dev *Device) Syspath() string {
	return dev.dev.Syspath()
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
func (dev *Device) OpenInterfaces(wr bool, ifaces wiimote.InterfaceKind) error {
	var errs []error
	for kind := wiimote.InterfaceCore; kind <= wiimote.InterfaceGuitar; kind <<= 1 {
		if ifaces&kind == 0 {
			continue
		}
		node, ok := dev.availIfs[kind]
		if !ok {
			continue
		}
		iface := InterfaceFromName(kind)
		if err := iface.open(dev, node, wr); err != nil {
			errs = append(errs, err)
			continue
		}
		dev.ifs[node] = iface
		dev.moreEvents <- &wiimote.EventInterface{
			Event: commonEvent{
				timestamp: time.Now(),
				iface:     iface,
			},
		}
	}
	return errors.Join(errs...)
}

// IsAvailable returns a bitmask of available devices. These devices can be opened and are
// guaranteed to be present on the hardware at this time. If you watch your
// device for hotplug events you will get notified whenever this bitmask changes.
// See the WatchEvent event for more information.
func (dev *Device) Available(iface wiimote.InterfaceKind) bool {
	_, ok := dev.availIfs[iface]
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
func (dev *Device) Poll() (wiimote.Event, bool, error) {
	select {
	case e := <-dev.moreEvents:
		return e, true, nil
	default:
	}

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

	return nil, false, common.ErrPollAgain
}

// LED reads the LED state for the given LED.
//
// LEDs are a static interface that does not have to be opened first.
func (dev *Device) LED() (result wiimote.Led, _ error) {
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
func (dev *Device) SetLED(leds wiimote.Led) error {
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

// Battery reads the current battery capacity. The capacity is represented as percentage, thus the return value is an integer between 0 and 100.
//
// Batteries are a static interface that does not have to be opened first.
func (dev *Device) Battery() (uint, error) {
	cont, err := os.ReadFile(dev.batteryAttr)
	if err != nil {
		return 0, err
	}

	cap, err := strconv.Atoi(strings.TrimSpace(string(cont)))
	return uint(cap), err
}

// DevType returns the device type. If the device type cannot be determined,
// it returns "unknown" and the corresponding error.
//
// This is a static interface that does not have to be opened first.
func (dev *Device) DevType() (string, error) {
	cont, err := os.ReadFile(dev.devtypeAttr)
	return strings.TrimSpace(string(cont)), err
}

// Extension returns the extension type. If no extension is connected or the
// extension cannot be determined, it returns a string "none" and the corresponding error.
//
// This is a static interface that does not have to be opened first.
func (dev *Device) Extension() (string, error) {
	cont, err := os.ReadFile(dev.extensionAttr)
	return strings.TrimSpace(string(cont)), err
}

func (dev *Device) String() string {
	var w strings.Builder
	w.WriteString("wiimote-device ")
	devtype, _ := dev.DevType()
	w.WriteString(devtype)
	ext, _ := dev.Extension()
	if ext != "none" && ext != "" {
		w.WriteString(" with ")
		w.WriteString(ext)
	}
	w.WriteString(" at ")
	w.WriteString(dev.Syspath())
	return w.String()
}
