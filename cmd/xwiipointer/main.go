package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/virtpointer"
	"github.com/friedelschoen/wayland"
)

func watchDevice(path string) {
	mouse, err := virtpointer.NewVirtualPointer()
	if err != nil {
		log.Fatalf("error: unable to create mouse: %v", err)
	}
	defer mouse.Close()

	dev, err := xwiimote.NewDevice(path)
	if err != nil {
		log.Fatalf("error: unable to get device: %v", err)
	}
	defer dev.Free()

	if err := dev.Open(xwiimote.InterfaceCore | xwiimote.InterfaceIR | xwiimote.InterfaceAccel); err != nil {
		log.Fatalf("error: unable to open device: %v", err)
	}

	bat, _ := dev.GetBattery()
	fmt.Printf("new wiimote at %s with %d%% battery\n", dev.GetSyspath(), bat)

	pointer := xwiimote.NewIRPointer()
	var lastIR *xwiimote.EventIR
	var lastAccel *xwiimote.EventAccel
	for {
		ev, err := dev.Wait(-1)
		if err != nil {
			log.Printf("unable to poll event: %v\n", err)
		}
		switch ev := ev.(type) {
		case *xwiimote.EventIR:
			lastIR = ev
		case *xwiimote.EventAccel:
			lastAccel = ev
		case *xwiimote.EventKey:
			if ev.State == xwiimote.StateRepeated {
				break
			}
			switch ev.Code {
			case xwiimote.KeyA:
				mouse.Button(wayland.ButtonLeft, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyB:
				mouse.Button(wayland.ButtonRight, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyHome:
				mouse.Button(wayland.ButtonMiddle, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyUp:
				if ev.State == xwiimote.StatePressed {
					mouse.Scroll(false, -10)
				}
			case xwiimote.KeyDown:
				if ev.State == xwiimote.StatePressed {
					mouse.Scroll(false, 10)
				}
			case xwiimote.KeyLeft:
				if ev.State == xwiimote.StatePressed {
					mouse.Scroll(true, -10)
				}
			case xwiimote.KeyRight:
				if ev.State == xwiimote.StatePressed {
					mouse.Scroll(true, 10)
				}
			}
		}
		if lastIR != nil && lastAccel != nil {
			roll := math.Atan2(float64(lastAccel.Accel.X), float64(lastAccel.Accel.Z))
			pointer.Update(lastIR.Slots, roll)
			lastIR = nil
			lastAccel = nil
		}
		if pointer.Health >= xwiimote.IRGood && pointer.Smooth != nil {
			x := xwiimote.MapNumber(pointer.Smooth.X, -340, 340, 0, 1000, true)
			y := xwiimote.MapNumber(pointer.Smooth.Y, -92, 290, 0, 1000, true)
			fmt.Printf("[%v] pointer at (%.2f %.2f) at %.2fm distance -> mapping to (%.0f %.0f)\n", pointer.Health, pointer.Smooth.X, pointer.Smooth.Y, pointer.WiimoteDistance(), x, y)
			mouse.Set(uint32(x), uint32(y), 1000, 1000)
		}
	}
}

func main() {
	flag.Parse()

	monitor := xwiimote.NewMonitor(xwiimote.MonitorUdev)
	defer monitor.Free()

	for {
		path, err := monitor.Wait(-1)
		if err != nil || path == "" {
			log.Printf("error while polling: %v\n", err)
			continue
		}
		watchDevice(path)
	}
}
