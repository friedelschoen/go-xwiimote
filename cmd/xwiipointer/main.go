package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/irpointer"
	"github.com/friedelschoen/go-xwiimote/pkg/vinput"
)

var ScrollSpeed = flag.Float64("-scrollspeed", 1.0, "Set the scrollspeed")

func watchDevice(dev *xwiimote.Device) {
	mouse, err := vinput.CreateMouse("xwiimote-mouse",
		vinput.Range{Min: -340, Max: 340, Res: 72},
		vinput.Range{Min: -92, Max: 290, Res: 72})
	if err != nil {
		log.Fatalf("error: unable to create mouse: %v", err)
	}
	defer mouse.Close()

	if err := dev.OpenInterfaces(false, &xwiimote.InterfaceCore{}, &xwiimote.InterfaceIR{}, &xwiimote.InterfaceAccel{}); err != nil {
		log.Fatalf("error: unable to open device: %v", err)
	}

	bat, _ := dev.GetBattery()
	fmt.Printf("new wiimote at %s with %d%% battery\n", dev.GetSyspath(), bat)

	pointer := irpointer.NewIRPointer(nil, irpointer.FRect{Min: irpointer.FVec2{X: -340, Y: -92}, Max: irpointer.FVec2{X: 340, Y: 290}})
	var frame irpointer.Frame
	var scroll *irpointer.FVec2
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
			case xwiimote.KeyDown:
				if ev.State == xwiimote.StatePressed {
					if frame.Valid {
						scroll = &frame.Position
					} else {
						scroll = &irpointer.FVec2{}
					}
				} else {
					scroll = nil
				}

			case xwiimote.KeyLeft:
				mouse.Key(vinput.ButtonBack, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyRight:
				mouse.Key(vinput.ButtonForward, ev.State != xwiimote.StateReleased)
			}
		}
		if lastIR != nil && lastAccel != nil {
			frame = pointer.Step(lastIR.Slots, lastAccel.Accel)
			lastIR = nil
			lastAccel = nil
		}
		if frame.Health >= irpointer.IRGood {
			x, y := frame.Position.X, frame.Position.Y
			if scroll != nil {
				dx := x - scroll.X
				dy := y - scroll.Y
				if dx > -10 && dx < 10 {
					dx = 0
				}
				if dy > -10 && dy < 10 {
					dy = 0
				}

				mouse.Scroll(int32(*ScrollSpeed*dx), int32(*ScrollSpeed*dy))
			} else if frame.InBounds {
				fmt.Printf("[%v] pointer at (%.2f %.2f) at %.2fm distance\n", frame.Health, frame.Position.X, frame.Position.Y, frame.Distance)
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
		go watchDevice(dev)
	}
}
