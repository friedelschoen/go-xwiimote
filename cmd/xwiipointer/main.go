package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/irpointer"
	"github.com/friedelschoen/go-xwiimote/pkg/vinput"
)

var ScrollSpeed = flag.Float64("scrollspeed", -0.01, "Set the vertical scrollspeed")
var HorizScrollSpeed = flag.Float64("hscrollspeed", -0.01, "Set the horizontal scrollspeed")

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
	fmt.Printf("new wiimote at %s with %d%% battery, cap=%v\n", dev.GetSyspath(), bat, dev.Available(&xwiimote.InterfaceIR{}))

	pointer := irpointer.NewIRPointer()
	process := irpointer.FilterChain{
		irpointer.NewErrorFilter(),
		irpointer.NewGlitchFilter(),
		irpointer.NewOneEuroSmoothing(),
	}

	var frame irpointer.Frame
	var scroll *irpointer.FVec2
	var lastIR *xwiimote.EventIR
	var lastAccel *xwiimote.EventAccel
	var hold bool
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
			if ev.Code != xwiimote.KeyDown {
				hold = ev.State == xwiimote.StatePressed
			}
			switch ev.Code {
			case xwiimote.KeyA:
				mouse.Key(vinput.ButtonLeft, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyB:
				mouse.Key(vinput.ButtonRight, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyHome:
				mouse.Key(vinput.ButtonMiddle, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyLeft:
				mouse.Key(vinput.ButtonBack, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyRight:
				mouse.Key(vinput.ButtonForward, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyMinus:
				mouse.Key(vinput.KeyVolumedown, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyPlus:
				mouse.Key(vinput.KeyVolumeup, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyTwo:
				mouse.Key(vinput.KeyPlaypause, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyOne:
				mouse.Key(vinput.KeyNext, ev.State != xwiimote.StateReleased)
			case xwiimote.KeyDown:
				if ev.State == xwiimote.StatePressed {
					if frame.Valid {
						pos := frame.Position
						scroll = &pos
					} else {
						scroll = &irpointer.FVec2{}
					}
				} else {
					scroll = nil
				}
			}
		}
		if !hold && lastIR != nil && lastAccel != nil {
			frame = pointer.Step(lastIR.Slots, lastAccel.Accel)
			frame = process.Apply(frame)
			lastIR = nil
			lastAccel = nil
		}
		if frame.Valid && frame.Health >= irpointer.IRGood {
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

				fmt.Printf("[%v] scroll to (%.2f %.2f) at %.2fcm distance\n", frame.Health, dx, dy, frame.Distance)
				mouse.Scroll(int32(*HorizScrollSpeed*dx), int32(*ScrollSpeed*dy))
			} else {
				fmt.Printf("[%v] pointer at (%.2f %.2f) at %.2fcm distance\n", frame.Health, x, y, frame.Distance)
				mouse.Set(int32(x), int32(y))
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
