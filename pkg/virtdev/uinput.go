package virtdev

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type uinputConstructor struct {
	path string
	id   inputID
}

var defaultUinputConstructor = uinputConstructor{
	path: "/dev/uinput",
	id: inputID{
		Bustype: busUsb,
		Product: 0xdead,
		Vendor:  0xbeef,
		Version: 0,
	},
}

type UinputOption func(*uinputConstructor)

func WithUinputPath(path string) UinputOption {
	return func(uc *uinputConstructor) {
		uc.path = path
	}
}

func WithVendorProduct(vendor, product, version uint16) UinputOption {
	return func(uc *uinputConstructor) {
		uc.id.Vendor = vendor
		uc.id.Product = product
		uc.id.Version = version
	}
}

type uinputDevice struct {
	deviceFile *os.File
}

func createUinputDevice(path string) (uinputDevice, error) {
	deviceFile, err := os.OpenFile(path, os.O_WRONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return uinputDevice{}, fmt.Errorf("could not open device file: %w", err)
	}
	return uinputDevice{deviceFile}, err
}

func (dev *uinputDevice) register(code uintptr, events ...uintptr) error {
	for _, ev := range events {
		err := dev.ioctl(code, ev)
		if err != nil {
			defer dev.Close()
			return fmt.Errorf("invalid file handle returned from ioctl: %w", err)
		}
	}
	return nil
}

func toUinputName(uinputName *[uiMaxNameSize]byte, name string) error {
	if name == "" {
		return errors.New("device name may not be empty")
	}
	if len(name) > uiMaxNameSize {
		return fmt.Errorf("device name %s is too long (maximum of %d characters allowed)", name, uiMaxNameSize)
	}
	copy(uinputName[:], name)
	return nil
}

func (dev *uinputDevice) setup(name string, busid inputID) error {
	setup := uinputSetup{id: busid}
	err := toUinputName(&setup.name, name)
	if err != nil {
		return err
	}
	err = dev.ioctl(uiDevSetup, uintptr(unsafe.Pointer(&setup)))
	if err != nil {
		dev.Close()
		return fmt.Errorf("failed to create device: %w", err)
	}
	return nil
}
func (dev *uinputDevice) create() error {
	err := dev.ioctl(uiDevCreate, 0)
	time.Sleep(time.Millisecond * 200)
	return err
}

func (dev *uinputDevice) GetSyspath() (string, error) {
	sysInputDir := "/sys/devices/virtual/input/"
	var path [uiSysnameLen + 1]byte
	err := dev.ioctl(uiGetSysname, uintptr(unsafe.Pointer(&path[0])))
	n := bytes.IndexByte(path[:], 0)
	if n < 0 {
		n = len(path)
	}
	name := string(path[:n])
	sysInputDir += name
	return sysInputDir, err
}

func (dev *uinputDevice) emit(typ, code uint16, value int32) error {
	ev := inputEvent{
		Time:  syscall.Timeval{},
		Type:  typ,
		Code:  code,
		Value: value,
	}
	_, err := dev.deviceFile.Write(ev.buffer())
	if err != nil {
		return fmt.Errorf("writing %v structure to the device file failed: %w", typ, err)
	}
	return nil
}

func (dev *uinputDevice) sync() (err error) {
	return dev.emit(evSyn, synReport, 0)
}

func (dev *uinputDevice) Close() (err error) {
	err = dev.releaseDevice()
	if err != nil {
		return fmt.Errorf("failed to close device: %w", err)
	}
	return dev.deviceFile.Close()
}

func (dev *uinputDevice) releaseDevice() (err error) {
	return dev.ioctl(uiDevDestroy, uintptr(0))
}

func (dev *uinputDevice) ioctl(cmd, ptr uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, dev.deviceFile.Fd(), cmd, ptr)
	if err == 0 {
		return nil
	}
	return err
}

func (dev *uinputDevice) Key(key Key, press bool) error {
	var state int32
	if press {
		state = 1
	}
	err := dev.emit(evKey, uint16(key), state)
	if err != nil {
		return err
	}
	return dev.sync()
}
