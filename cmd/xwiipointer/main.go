package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/irpointer"
	"github.com/friedelschoen/go-xwiimote/pkg/vinput"
)

func watchDevice(dev *xwiimote.Device) {
	mouse, err := vinput.CreateMouse("xwiimote-mouse",
		vinput.Range{Min: -340, Max: 340, Res: 72},
		vinput.Range{Min: -92, Max: 290, Res: 72})
	if err != nil {
		log.Fatalf("error: unable to create mouse: %v", err)
	}
	defer mouse.Close()

	if err := dev.OpenInterfaces(xwiimote.InterfaceCore|xwiimote.InterfaceIR|xwiimote.InterfaceAccel, false); err != nil {
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
				fmt.Println("err: ", err)
			}
		}
	}
}

func main() {
	flag.Parse()

	monitor, err := xwiimote.NewMonitor(xwiimote.MonitorUdev)
	if err != nil {
		log.Fatalln("error: ", err)
	}

	for {
		dev, err := monitor.Wait(-1)
		if err != nil || dev == nil {
			log.Printf("error while polling: %v\n", err)
			continue
		}
		watchDevice(dev)
	}
}
