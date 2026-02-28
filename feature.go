package wiimote

import (
	"io"
)

type Feature interface {
	io.Closer
	// Type of this feature
	Kind() FeatureKind
	// Device is the parent of this feature. When opened, a device is bound to this feature.
	Device() Device
	// Opened returns a bitmask of opened features. Features may be closed due to
	// error-conditions at any time. However, features are never opened
	// automatically.
	//
	// You will get notified whenever this bitmask changes, except on explicit
	// calls to Open() and Close(). See the EventWatch event for more information.
	Opened() bool
}

type RumbleFeature interface {
	Feature

	// Rumble sets the rumble motor.
	//
	// This requires the core-feature to be opened in writable mode.
	Rumble(state bool) error
}

type MemoryFeature interface {
	Feature

	Memory() (Memory, error)
}

type MotionPlusFeature interface {
	Feature

	// SetMPNormalization sets Motion-Plus normalization and calibration values. The Motion-Plus sensor is very
	// sensitive and may return really crappy values. This features allows to
	// apply 3 absolute offsets x, y and z which are subtracted from any MP data
	// before it is returned to the application. That is, if you set these values
	// to 0, this has no effect (which is also the initial state).
	//
	// The calibration factor is used to perform runtime calibration. If
	// it is 0 (the initial state), no runtime calibration is performed. Otherwise,
	// the factor is used to re-calibrate the zero-point of MP data depending on MP
	// input. This is an angoing calibration which modifies the internal state of
	// the x, y and z values.
	SetMPNormalization(x, y, z, factor int32)

	// MPNormalization reads the Motion-Plus normalization and calibration values. Please see
	// SetMPNormalization() how this is handled.
	//
	// Note that if the calibration factor is not 0, the normalization values may
	// change depending on incoming MP data. Therefore, the data read via this
	// function may differ from the values that you wrote to previously. However,
	// apart from applied calibration, these value are the same as were set
	// previously via SetMPNormalization() and you can feed them back
	// in later.
	MPNormalization() (x, y, z, factor int32)
}

type FeatureKind uint

const (
	FeatureCore FeatureKind = 1 << iota
	FeatureAccel
	FeatureIR
	FeatureSpeaker
	FeatureMotionPlus
	FeatureNunchuck
	FeatureClassicController
	FeatureBalanceBoard
	FeatureProController
	FeatureDrums
	FeatureGuitar

	FeatureSetCore = FeatureCore | FeatureAccel | FeatureIR | FeatureSpeaker
)
