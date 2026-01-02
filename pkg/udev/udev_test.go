package udev

import (
	"fmt"
	"testing"
)

func ExampleUdev_NewDeviceFromDevnum() {
	d := NewDeviceFromDevnum('c', MkDev(1, 8))
	fmt.Println(d.Syspath())
	// Output:
	// /sys/devices/virtual/mem/random
}

func TestNewDeviceFromDevnum(t *testing.T) {
	d := NewDeviceFromDevnum('c', MkDev(1, 8))
	if d.Devnum().Major() != 1 {
		t.Fail()
	}
	if d.Devnum().Minor() != 8 {
		t.Fail()
	}
	if d.Devpath() != "/devices/virtual/mem/random" {
		t.Fail()
	}
}

func ExampleUdev_NewDeviceFromSyspath() {
	d := NewDeviceFromSyspath("/sys/devices/virtual/mem/random")
	fmt.Println(d.Syspath())
	// Output:
	// /sys/devices/virtual/mem/random
}

func TestNewDeviceFromSyspath(t *testing.T) {
	d := NewDeviceFromSyspath("/sys/devices/virtual/mem/random")
	if d.Devnum().Major() != 1 {
		t.Fail()
	}
	if d.Devnum().Minor() != 8 {
		t.Fail()
	}
	if d.Devpath() != "/devices/virtual/mem/random" {
		t.Fail()
	}
}

func ExampleUdev_NewDeviceFromSubsystemSysname() {
	d := NewDeviceFromSubsystemSysname("mem", "random")
	fmt.Println(d.Syspath())
	// Output:
	// /sys/devices/virtual/mem/random
}

func TestNewDeviceFromSubsystemSysname(t *testing.T) {
	d := NewDeviceFromSubsystemSysname("mem", "random")
	if d.Devnum().Major() != 1 {
		t.Fail()
	}
	if d.Devnum().Minor() != 8 {
		t.Fail()
	}
	if d.Devpath() != "/devices/virtual/mem/random" {
		t.Fail()
	}
}

func ExampleUdev_NewDeviceFromDeviceID() {
	d := NewDeviceFromDeviceID("c1:8")
	fmt.Println(d.Syspath())
	// Output:
	// /sys/devices/virtual/mem/random
}

func TestNewDeviceFromDeviceID(t *testing.T) {
	d := NewDeviceFromDeviceID("c1:8")
	if d.Devnum().Major() != 1 {
		t.Fail()
	}
	if d.Devnum().Minor() != 8 {
		t.Fail()
	}
	if d.Devpath() != "/devices/virtual/mem/random" {
		t.Fail()
	}
}

func ExampleUdev_NewMonitorFromNetlink() {
	_ = NewMonitorFromNetlink("udev")
}

func TestNewMonitorFromNetlink(t *testing.T) {
	_ = NewMonitorFromNetlink("udev")
}

func ExampleUdev_NewEnumerate() {
	_ = NewEnumerate()
}

func TestNewEnumerate(t *testing.T) {
	_ = NewEnumerate()
}
