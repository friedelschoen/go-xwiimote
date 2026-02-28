package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/friedelschoen/go-uinput"
	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/driver"
	"github.com/friedelschoen/go-wiimote/pkg/discover"
)

var (
	kbname = flag.String("name", "wiimote-virtual", "Name to use")
)

func loadMapping(r io.Reader) map[wiimote.Key]uinput.Key {
	mapping := make(map[wiimote.Key]uinput.Key)
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		line := scan.Text()
		wiibuttonstr, realkeystr, ok := strings.Cut(line, "->")
		if !ok {
			fmt.Fprintf(os.Stderr, "error: missing delimiter: %s\n", line)
			continue
		}
		wiibutton, ok := wiimote.LookupKey(strings.TrimSpace(wiibuttonstr))
		if !ok {
			fmt.Fprintf(os.Stderr, "error: unknown button: %s\n", wiibuttonstr)
			continue
		}
		realkey, ok := uinput.LookupKey(strings.TrimSpace(realkeystr))
		if !ok {
			fmt.Fprintf(os.Stderr, "error: unknown key: %s\n", realkeystr)
			continue
		}
		mapping[wiibutton] = realkey
	}
	return mapping
}

func watchDevice(dev wiimote.Device, mapping map[wiimote.Key]uinput.Key) {
	fmt.Printf("new device: %s\n", dev.String())
	time.Sleep(100 * time.Millisecond)
	if err := dev.OpenFeatures(wiimote.FeatureCore, true); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open device: %s", err)
	}

	kb, err := uinput.CreateKeyboard(*kbname)
	if err != nil {
		panic(err)
	}
	defer kb.Close()
	var leds wiimote.Led

	rumbleif := dev.Feature(wiimote.FeatureCore).(wiimote.RumbleFeature)
	for {
		ev, err := dev.Wait(-1)
		if err != nil {
			log.Printf("unable to poll event: %v\n", err)
		}
		switch ev := ev.(type) {
		case *wiimote.EventKey:
			if ev.Code == wiimote.KeyHome {
				if rumbleif != nil {
					rumbleif.Rumble(ev.Pressed)
				}
				continue
			} else if ev.Code == wiimote.KeyTwo {
				if ev.Pressed {
					leds++
					leds %= 16

					fmt.Println(dev.SetLED(leds))
					continue
				}
			}

			realkey, ok := mapping[ev.Code]
			if !ok {
				continue
			}
			kb.Key(realkey, ev.Pressed)
		case *wiimote.EventGone:
			return
		}
	}
}

func main() {
	flag.Parse()

	mapping := loadMapping(os.Stdin)

	monitor, err := discover.NewWiimoteMonitor()
	if err != nil {
		log.Fatalln("error: ", err)
	}

	fmt.Println("waiting for devices...")
	for {
		dev, err := monitor.Wait(-1)
		if err != nil || dev == nil {
			log.Printf("error while polling: %v\n", err)
			continue
		}
		d, err := driver.NewDevice(dev, driver.BackendKernel)
		if err != nil {
			log.Printf("error creating device: %v\n", err)
			continue
		}
		watchDevice(d, mapping)
	}
}
