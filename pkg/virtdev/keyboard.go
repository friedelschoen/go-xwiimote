package virtdev

import (
	"fmt"
)

type Keyboard struct {
	uinputDevice
	name string
}

// CreateKeyboard will create a new keyboard using the given uinput
// device path of the uinput device.
func CreateKeyboard(name string, opts ...UinputOption) (*Keyboard, error) {
	construct := defaultUinputConstructor
	for _, opt := range opts {
		opt(&construct)
	}
	dev, err := createUinputDevice(construct.path)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual keyboard device: %w", err)
	}

	err = dev.register(uiSetEvBit, evKey, evSyn)
	if err != nil {
		dev.Close()
		return nil, fmt.Errorf("failed to register virtual keyboard device: %w", err)
	}

	// register key events
	for i := KeyReserved; i <= KeyMax; i++ {
		err = dev.register(uiSetKeyBit, uintptr(i))
		if err != nil {
			dev.Close()
			return nil, fmt.Errorf("failed to register key number %d: %v", i, err)
		}
	}

	err = dev.setup(name, construct.id)
	if err != nil {
		return nil, err
	}
	err = dev.create()

	return &Keyboard{dev, name}, err
}
