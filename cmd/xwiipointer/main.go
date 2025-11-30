package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/virtdev"
)

func watchDevice(path string) {
	mouse, err := virtdev.CreateMouse("xwiimote-mouse",
		virtdev.Range{Min: -340, Max: 340, Res: 72},
		virtdev.Range{Min: -92, Max: 290, Res: 72})
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
				mouse.Key(virtdev.ButtonLeft, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyB:
				mouse.Key(virtdev.ButtonRight, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyHome:
				mouse.Key(virtdev.ButtonMiddle, ev.State != xwiimote.StateReleased)
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
			pointer.Update(lastIR.Slots, lastAccel.Accel)
			lastIR = nil
			lastAccel = nil
		}
		if pointer.Health >= xwiimote.IRGood && pointer.Smooth != nil {
			x, y := pointer.Smooth.X, pointer.Smooth.Y
			if x >= -340 && x < 340 && y >= -92 && y < 290 {
				fmt.Printf("[%v] pointer at (%.2f %.2f) at %.2fm distance\n", pointer.Health, pointer.Smooth.X, pointer.Smooth.Y, pointer.Distance)
				err := mouse.Set(int32(x), int32(y))
				fmt.Println(err)
			}
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
