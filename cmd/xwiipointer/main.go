package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/irpointer"
	"github.com/friedelschoen/go-xwiimote/pkg/vinput"
)

var ScrollSpeed = flag.Float64("scrollspeed", 0.01, "Set the vertical scrollspeed")
var HorizScrollSpeed = flag.Float64("hscrollspeed", 0.01, "Set the horizontal scrollspeed")

func watchDevice(dev *xwiimote.Device) {
	mouse, err := vinput.CreateMouse("xwiimote-mouse",
		vinput.Range{Min: -340, Max: 340, Res: 72},
		vinput.Range{Min: -92, Max: 290, Res: 72}, []vinput.Key{
			vinput.ButtonLeft,
			vinput.ButtonRight,
			vinput.KeyLeftmeta,
			vinput.ButtonBack,
			vinput.ButtonForward,
			vinput.KeyVolumedown,
			vinput.KeyVolumeup,
			vinput.KeyPlaypause,
			vinput.KeyNext,
		})
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
		irpointer.NewRepeatFilter(),
	}

	holdSmooth := irpointer.NewOneEuroSmoothing()
	holdSmooth.MinCutoff = 0.15
	holdSmooth.Beta = 0.005
	holdSmooth.DCutoff = 0.8
	holdProcess := irpointer.FilterChain{
		irpointer.NewErrorFilter(),
		irpointer.NewGlitchFilter(),
		holdSmooth,
		irpointer.NewRepeatFilter(),
	}

	var frame irpointer.Frame
	go func() {
		blink := false
		lostcount := 0
		for {
			blink = !blink

			var leds xwiimote.Led
			switch frame.Health {
			case irpointer.IRLost:
				if blink && lostcount <= 5 {
					leds |= xwiimote.Led1
					lostcount++
				}
			case irpointer.IRSingle:
				leds |= xwiimote.Led1
				lostcount = 0
			case irpointer.IRGood:
				leds |= xwiimote.Led1 | xwiimote.Led2
				lostcount = 0
			}

			soc, err := dev.GetBattery()
			if err != nil {
				/* assume full */
				soc = 100
			}
			switch {
			case soc < 5:
				if blink {
					leds |= xwiimote.Led3
				}
			case soc < 25:
				leds |= xwiimote.Led3
			default:
				leds |= xwiimote.Led3 | xwiimote.Led4
			}

			dev.SetLED(leds)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	var scroll *irpointer.FVec2
	go func() {
		for {
			if scroll == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			dx := scroll.X - frame.Position.X
			dy := scroll.Y - frame.Position.Y
			scrollx, scrolly := int32(*HorizScrollSpeed*dx), int32(*ScrollSpeed*dy)
			fmt.Printf("[%v] scroll to (%d %d) at %.2fcm distance\n", frame.Health, scrollx, scrolly, frame.Distance)
			mouse.Scroll(scrollx, scrolly)
			time.Sleep(50 * time.Millisecond)
		}
	}()

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
				mouse.Key(vinput.KeyLeftmeta, ev.State != xwiimote.StateReleased)
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
		if lastIR != nil && lastAccel != nil {
			frame = pointer.Step(lastIR.Slots, lastAccel.Accel)
			if hold {
				frame = holdProcess.Apply(frame)
			} else {
				frame = process.Apply(frame)
			}
			lastIR = nil
			lastAccel = nil
		}
		if frame.Valid && frame.Health >= irpointer.IRGood {
			x, y := frame.Position.X, frame.Position.Y
			if scroll == nil {
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
