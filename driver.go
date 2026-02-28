package wiimote

//go:generate morestringer -lookup Lookup{} -output stringer.go Led Key:cconst FeatureKind

import (
	"fmt"
	"io"
	"time"
)

type Memory interface {
	io.WriterAt
	io.ReaderAt
	io.Closer
}

// Led described a Led of an device. The leds are counted left-to-right and can be OR'ed together.
type Led uint8

const (
	Led1 Led = 1 << iota
	Led2
	Led3
	Led4
)

type Device interface {
	fmt.Stringer
	Poller[Event]

	// Syspath returns the sysfs path of the underlying device. It is not neccesarily
	// the same as the one during NewDevice. However, it is guaranteed to
	// point at the same device (symlinks may be resolved).
	Syspath() string

	// OpenFeatures all the requested features. If FeatureWritable is also set,
	// the features are opened with write-access. Note that features that are
	// already opened are ignored and not touched.
	// If any feature fails to open, this function still tries to open the other
	// requested features and then returns the error afterwards. Hence, if this
	// function fails, you should use Opened() to get a bitmask of opened
	// features and see which failed (if that is of interest).
	//
	// Note that features may be closed automatically during runtime if the
	// kernel removes the feature or on error conditions. You always get an
	// EventWatch event which you should react on. This is returned
	// regardless whether Watch() was enabled or not.
	OpenFeatures(ifaces FeatureKind, wr bool) error

	// Feature receives an feature and returns nil this feature is not opened
	Feature(ifaces FeatureKind) Feature

	// IsAvailable returns a bitmask of available devices. These devices can be opened and are
	// guaranteed to be present on the hardware at this time. If you watch your
	// device for hotplug events you will get notified whenever this bitmask changes.
	// See the WatchEvent event for more information.
	Available(iface FeatureKind) bool

	// SetIRFull sets
	IRFull() bool

	// SetIRFull sets
	SetIRFull(fullreport bool)

	// LED reads the LED state for the given LED.
	//
	// LEDs are a static feature that does not have to be opened first.
	LED() (result Led, _ error)

	// SetLED writes the LED state for the given LED.
	//
	// LEDs are a static feature that does not have to be opened first.
	SetLED(leds Led) error

	// Battery reads the current battery capacity. The capacity is represented as percentage, thus the return value is an integer between 0 and 100.
	//
	// Batteries are a static feature that does not have to be opened first.
	Battery() (uint, error)

	// DevType returns the device type. If the device type cannot be determined,
	// it returns "unknown" and the corresponding error.
	//
	// This is a static feature that does not have to be opened first.
	DevType() (string, error)

	// Extension returns the extension type. If no extension is connected or the
	// extension cannot be determined, it returns a string "none" and the corresponding error.
	//
	// This is a static feature that does not have to be opened first.
	Extension() (string, error)
}

type Poller[T any] interface {
	// Poll attempts to retrieve an event or data.
	//
	// Return values:
	//   T:     the retrieved data (invalid if error == ErrRetry)
	//   bool:  indicates whether more data is immediately available without
	//          waiting for I/O readiness
	//   error: nil on success. If ErrRetry is returned, the call should be
	//          repeated without waiting. Any other error aborts the attempt.
	Poll() (T, bool, error)

	// Wait waits for an event up to the specified timeout. A negative timeout is considered forever.
	// It handles ErrRetry internally and returns the first valid event or error.
	Wait(timeout time.Duration) (T, error)

	// Wait waits for an event up to the specified timeout. A negative timeout is considered forever.
	WaitReadable(timeout time.Duration) error

	// Handle continuously polls and calls `yield` with new events.
	// It blocks forever and should be used in a new goroutine.
	Handle(yield func(T)) error

	// Stream continuously polls and writes events into ch. It is a wrapper for Handle.
	// It blocks forever and should be used in a new goroutine.
	//
	//	p.Handle(func(ev T) { ch <- ev })
	Stream(ch chan<- T)
}
