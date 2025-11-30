package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/irpointer"
	"github.com/friedelschoen/go-xwiimote/pkg/vinput"
)

func watchDevice(path string) {
	mouse, err := vinput.CreateMouse("xwiimote-mouse",
		vinput.Range{Min: -340, Max: 340, Res: 72},
		vinput.Range{Min: -92, Max: 290, Res: 72})
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

	pointer := irpointer.NewIRPointer(nil)
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
				mouse.Key(vinput.ButtonLeft, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyB:
				mouse.Key(vinput.ButtonRight, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyHome:
				mouse.Key(vinput.ButtonMiddle, ev.State != xwiimote.StateReleased)
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
		if pointer.Health >= irpointer.IRGood && pointer.Position != nil {
			x, y := pointer.Position.X, pointer.Position.Y
			if x >= -340 && x < 340 && y >= -92 && y < 290 {
				fmt.Printf("[%v] pointer at (%.2f %.2f) at %.2fm distance\n", pointer.Health, pointer.Position.X, pointer.Position.Y, pointer.Distance)
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
