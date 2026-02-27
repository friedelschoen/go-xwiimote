package udev

import (
	"fmt"
	"testing"
)

func ExampleNewDeviceFromSyspath() {
	d := NewDeviceFromSyspath("/sys/devices/virtual/mem/random")
	fmt.Println(d.Syspath())
	// Output:
	// /sys/devices/virtual/mem/random
}

func ExampleNewDeviceFromSubsystemSysname() {
	d := NewDeviceFromSubsystemSysname("mem", "random")
	fmt.Println(d.Syspath())
	// Output:
	// /sys/devices/virtual/mem/random
}

func ExampleNewDeviceFromDeviceID() {
	d := NewDeviceFromDeviceID("c1:8")
	fmt.Println(d.Syspath())
	// Output:
	// /sys/devices/virtual/mem/random
}

func ExampleNewMonitorFromNetlink() {
	_ = NewMonitorFromNetlink(MonitorUdev)
}

func TestNewMonitorFromNetlink(t *testing.T) {
	_ = NewMonitorFromNetlink(MonitorUdev)
}

func ExampleNewEnumerate() {
	_ = NewEnumerate()
}

func TestNewEnumerate(t *testing.T) {
	_ = NewEnumerate()
}
