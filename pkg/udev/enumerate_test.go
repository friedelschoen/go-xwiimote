package udev

import (
	"fmt"
	"runtime"
	"slices"
	"testing"
)

func TestEnumerateDeviceSyspaths(t *testing.T) {
	e := NewEnumerate()
	dsp, err := e.Devices()
	if err != nil {
		t.Fail()
	}
	if len(slices.Collect(dsp)) <= 0 {
		t.Fail()
	}
}
func ExampleEnumerate_SubsystemSyspaths() {
	e := NewEnumerate()

	// Enumerate all subsystem syspaths
	dsp, _ := e.SubsystemSyspaths()
	for s := range dsp {
		fmt.Println(s)
	}
}

func TestEnumerateSubsystemSyspaths(t *testing.T) {
	e := NewEnumerate()
	ssp, err := e.SubsystemSyspaths()
	if err != nil {
		t.Fail()
	}
	if len(slices.Collect(ssp)) == 0 {
		t.Fail()
	}
}

func ExampleEnumerate_Devices() {
	e := NewEnumerate()

	// Add some FilterAddMatchSubsystemDevtype
	e.AddMatchSubsystem("block")
	e.AddMatchIsInitialized()
	devices, _ := e.Devices()
	for device := range devices {
		fmt.Println(device)
	}
}

func TestEnumerateDevicesWithFilter(t *testing.T) {
	e := NewEnumerate()
	e.AddMatchSubsystem("block")
	e.AddMatchIsInitialized()
	e.AddNomatchSubsystem("mem")
	e.AddMatchProperty("ID_TYPE", "disk")
	e.AddMatchSysattr("partition", "1")
	e.AddMatchTag("systemd")
	//	e.AddMatchProperty("DEVTYPE", "partition")
	ds, err := e.Devices()
	if err != nil {
		t.Fail()
	}
	for path := range ds {
		d := NewDeviceFromSyspath(path)
		if d.Subsystem() != "block" {
			t.Error("Wrong subsystem")
		}
		if !d.IsInitialized() {
			t.Error("Not initialized")
		}
		if d.PropertyValue("ID_TYPE") != "disk" {
			t.Error("Wrong ID_TYPE")
		}
		if d.SysattrValue("partition") != "1" {
			t.Error("Wrong partition")
		}
		if !d.HasTag("systemd") {
			t.Error("Not tagged")
		}

		parent := d.Parent()
		parent2 := d.ParentWithSubsystemDevtype("block", "disk")
		if parent.Syspath() != parent2.Syspath() {
			t.Error("Parent syspaths don't match")
		}

	}
}

func TestEnumerateGC(t *testing.T) {
	runtime.GC()
}
