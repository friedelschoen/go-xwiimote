package vinput

import (
	"errors"
	"fmt"
	"unsafe"
)

// A Mouse is a device that will trigger an absolute change event.
// For details see: https://www.kernel.org/doc/Documentation/input/event-codes.txt
type Mouse struct {
	uinputDevice
	name string
}

type Range struct {
	Min, Max, Res int
}

func (rng Range) empty() bool {
	return rng.Min == rng.Max
}

// CreateMouse will create a new mouse input device. A mouse is a device that allows relative input.
// Relative input means that all changes to the x and y coordinates of the mouse pointer will be
func CreateMouse(name string, xabs, yabs Range, opts ...UinputOption) (*Mouse, error) {
	construct := defaultUinputConstructor
	for _, opt := range opts {
		opt(&construct)
	}
	devfile, err := createUinputDevice(construct.path)
	if err != nil {
		return nil, fmt.Errorf("could not create relative axis input device: %w", err)
	}
	dev := &Mouse{devfile, name}

	err = dev.register(uiSetEvBit, evSyn, evKey, evRel, evAbs)
	if err != nil {
		dev.Close()
		return nil, fmt.Errorf("failed to register key device: %w", err)
	}

	// register button events (in order to enable left, right and middle click)
	err = dev.register(uiSetKeyBit, uintptr(ButtonLeft), uintptr(ButtonRight), uintptr(ButtonMiddle))
	if err != nil {
		dev.Close()
		return nil, fmt.Errorf("failed to register key: %w", err)
	}

	err = dev.register(uiSetRelBit, relX, relY, relWheel, relHWheel)
	if err != nil {
		dev.Close()
		return nil, fmt.Errorf("failed to register relative event: %w", err)
	}

	err = dev.setup(name, construct.id)
	if err != nil {
		dev.Close()
		return nil, fmt.Errorf("failed to create usb: %w", err)
	}

	if !xabs.empty() {
		if err := dev.registerAbs(absX, int32(xabs.Min), int32(xabs.Max), int32(xabs.Res)); err != nil {
			dev.Close()
			return nil, fmt.Errorf("failed to register absolute event: %w", err)
		}
	}
	if !yabs.empty() {
		if err := dev.registerAbs(absY, int32(yabs.Min), int32(yabs.Max), int32(yabs.Res)); err != nil {
			dev.Close()
			return nil, fmt.Errorf("failed to register absolute event: %w", err)
		}
	}

	err = dev.create()
	return dev, err
}

// Move will perform a move of the mouse pointer along the x and y axes relative to the current position as requested.
// Note that the upper left corner is (0, 0), so positive x and y means moving right (x) and down (y), whereas negative
// values will cause a move towards the upper left corner.
func (vRel *Mouse) Move(x, y int32) error {
	if err := vRel.emit(evRel, relX, x); err != nil {
		return fmt.Errorf("failed to move pointer along x axis: %w", err)
	}
	if err := vRel.emit(evRel, relY, y); err != nil {
		return fmt.Errorf("failed to move pointer along y axis: %w", err)
	}
	vRel.sync()
	return nil
}

func (vRel *Mouse) Set(x, y int32) error {
	if err := vRel.emit(evAbs, absX, x); err != nil {
		return fmt.Errorf("failed to move pointer along x axis: %w", err)
	}
	if err := vRel.emit(evAbs, absY, y); err != nil {
		return fmt.Errorf("failed to move pointer along y axis: %w", err)
	}
	vRel.sync()
	return nil
}

// Scroll will simulate a wheel movement.
func (vRel *Mouse) Scroll(x, y int32) error {
	var errs [2]error
	if x != 0 {
		errs[0] = vRel.emit(evRel, relHWheel, x)
	}
	if y != 0 {
		errs[1] = vRel.emit(evRel, relHWheel, y)
	}
	if err := errors.Join(errs[:]...); err != nil {
		return err
	}
	return vRel.sync()
}

func (vRel *Mouse) registerAbs(typ uint16, min, max, res int32) error {
	s := absSetup{
		code: typ,
		absinfo: absInfo{
			minimum:    min,
			maximum:    max,
			resolution: res,
		},
	}
	return vRel.ioctl(uiAbsSetup, uintptr(unsafe.Pointer(&s)))
}
