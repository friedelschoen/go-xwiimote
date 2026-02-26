package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/pkg/irpointer"
	"github.com/friedelschoen/go-wiimote/pkg/vinput"
)

var ScrollSpeed = flag.Float64("scrollspeed", 0.01, "Set the vertical scrollspeed")
var HorizScrollSpeed = flag.Float64("hscrollspeed", 0.01, "Set the horizontal scrollspeed")

func watchDevice(dev *wiimote.Device) {
	mouse, err := vinput.CreateMouse("wiimote-mouse",
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

	if err := dev.OpenInterfaces(false, &wiimote.InterfaceCore{}, &wiimote.InterfaceIR{}, &wiimote.InterfaceAccel{}); err != nil {
		log.Fatalf("error: unable to open device: %v", err)
	}

	bat, _ := dev.Battery()
	fmt.Printf("new wiimote at %s with %d%% battery, cap=%v\n", dev.Syspath(), bat, dev.Available(&wiimote.InterfaceIR{}))

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

			var leds wiimote.Led
			switch frame.Health {
			case irpointer.IRLost:
				if blink && lostcount <= 5 {
					leds |= wiimote.Led1
					lostcount++
				}
			case irpointer.IRSingle:
				leds |= wiimote.Led1
				lostcount = 0
			case irpointer.IRGood:
				leds |= wiimote.Led1 | wiimote.Led2
				lostcount = 0
			}

			soc, err := dev.Battery()
			if err != nil {
				/* assume full */
				soc = 100
			}
			switch {
			case soc < 5:
				if blink {
					leds |= wiimote.Led3
				}
			case soc < 25:
				leds |= wiimote.Led3
			default:
				leds |= wiimote.Led3 | wiimote.Led4
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

	var lastIR *wiimote.EventIR
	var lastAccel *wiimote.EventAccel
	var hold time.Time
	for {
		ev, err := dev.Wait(-1)
		if err != nil {
			log.Printf("unable to poll event: %v\n", err)
		}
		switch ev := ev.(type) {
		case *wiimote.EventIR:
			lastIR = ev
		case *wiimote.EventAccel:
			lastAccel = ev
		case *wiimote.EventKey:
			if ev.State == wiimote.StateRepeated {
				break
			}
			if ev.Code != wiimote.KeyDown {
				hold = time.Time{}
				if ev.State == wiimote.StatePressed {
					hold = time.Now()
				}
			}
			switch ev.Code {
			case wiimote.KeyA:
				mouse.Key(vinput.ButtonLeft, ev.State != wiimote.StateReleased)
			case wiimote.KeyB:
				mouse.Key(vinput.ButtonRight, ev.State != wiimote.StateReleased)
			case wiimote.KeyHome:
				mouse.Key(vinput.KeyLeftmeta, ev.State != wiimote.StateReleased)
			case wiimote.KeyLeft:
				mouse.Key(vinput.ButtonBack, ev.State != wiimote.StateReleased)
			case wiimote.KeyRight:
				mouse.Key(vinput.ButtonForward, ev.State != wiimote.StateReleased)
			case wiimote.KeyMinus:
				mouse.Key(vinput.KeyVolumedown, ev.State != wiimote.StateReleased)
			case wiimote.KeyPlus:
				mouse.Key(vinput.KeyVolumeup, ev.State != wiimote.StateReleased)
			case wiimote.KeyTwo:
				mouse.Key(vinput.KeyPlaypause, ev.State != wiimote.StateReleased)
			case wiimote.KeyOne:
				mouse.Key(vinput.KeyNext, ev.State != wiimote.StateReleased)
			case wiimote.KeyDown:
				if ev.State == wiimote.StatePressed {
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
		if lastIR != nil && lastAccel != nil && (hold.IsZero() || time.Since(hold) > 500*time.Millisecond) {
			frame = pointer.Step(lastIR.Slots, lastAccel.Accel)
			holdframe := holdProcess.Apply(frame)
			regframe := process.Apply(frame)
			if !hold.IsZero() {
				frame = holdframe
			} else {
				frame = regframe
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

	monitor, err := wiimote.NewMonitor(wiimote.MonitorUdev)
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
