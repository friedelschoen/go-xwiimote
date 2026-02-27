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
func ExampleEnumerate_Subsystems() {
	e := NewEnumerate()

	// Enumerate all subsystem syspaths
	dsp, _ := e.Subsystems()
	for s := range dsp {
		fmt.Println(s)
	}
}

func TestEnumerateSubsystemSyspaths(t *testing.T) {
	e := NewEnumerate()
	ssp, err := e.Subsystems()
	if err != nil {
		t.Fail()
	}
	if len(slices.Collect(ssp)) == 0 {
		t.Fail()
	}
}

func TestEnumerateGC(t *testing.T) {
	runtime.GC()
}
