package udev

import (
	"fmt"
	"runtime"
	"testing"
)

func ExampleDevice() {
	// Create new Device based on subsystem and sysname
	d := NewDeviceFromSubsystemSysname("mem", "zero")

	// Extract information
	fmt.Printf("Sysname:%v\n", d.Sysname())
	fmt.Printf("Syspath:%v\n", d.Syspath())
	fmt.Printf("Devnode:%v\n", d.Devnode())
	fmt.Printf("Subsystem:%v\n", d.Subsystem())
	fmt.Printf("Driver:%v\n", d.Driver())

	// Output:
	// Sysname:zero
	// Syspath:/sys/devices/virtual/mem/zero
	// Devpath:/devices/virtual/mem/zero
	// Devnode:/dev/zero
	// Subsystem:mem
	// Devtype:
	// Sysnum:
	// Driver:
}

func TestDeviceZero(t *testing.T) {
	d := NewDeviceFromDeviceID("c1:5")
	if d.Subsystem() != "mem" {
		t.Fail()
	}
	if d.Sysname() != "zero" {
		t.Fail()
	}
	if d.Syspath() != "/sys/devices/virtual/mem/zero" {
		t.Fail()
	}
	if d.Devnode() != "/dev/zero" {
		t.Fail()
	}
	if d.SysattrValue("subsystem") != "mem" {
		t.Fail()
	}
}

func TestDeviceGC(t *testing.T) {
	runtime.GC()
}
