package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/friedelschoen/go-wiimote"
)

var (
	openIf = flag.String("interfaces", "", "interfaces to use")
)

type eventBlock struct {
	Type      string        `json:"type"`
	Event     wiimote.Event `json:"event"`
	Id        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Interface string        `json:"interface"`
}

func watchDevice(dev *wiimote.Device, mu *sync.Mutex) {
	fmt.Printf("new device: %s\n", dev.String())
	time.Sleep(100 * time.Millisecond)
	var ifs []wiimote.Interface
	ifs = append(ifs, &wiimote.InterfaceCore{})
	for name := range strings.SplitSeq(*openIf, ",") {
		switch name {
		case "accel":
			ifs = append(ifs, &wiimote.InterfaceAccel{})
		case "bb", "balanceboard":
			ifs = append(ifs, &wiimote.InterfaceBalanceBoard{})
		case "cc", "classiccontroller":
			ifs = append(ifs, &wiimote.InterfaceClassicController{})
		case "drums":
			ifs = append(ifs, &wiimote.InterfaceDrums{})
		case "guitar":
			ifs = append(ifs, &wiimote.InterfaceGuitar{})
		case "ir":
			ifs = append(ifs, &wiimote.InterfaceIR{})
		case "mp", "motionplus":
			ifs = append(ifs, &wiimote.InterfaceMotionPlus{})
		case "nunchuck":
			ifs = append(ifs, &wiimote.InterfaceNunchuck{})
		case "pc", "procontroller":
			ifs = append(ifs, &wiimote.InterfaceProController{})
		}
	}
	if err := dev.OpenInterfaces(true, ifs...); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open device: %s", err)
	}
	dev.Watch(true)
	var block eventBlock
	for {
		ev, err := dev.Wait(-1)
		if err != nil {
			log.Printf("unable to poll event: %v\n", err)
		}
		if _, ok := ev.(*wiimote.EventGone); ok {
			return
		}

		block.Type = fmt.Sprintf("%T", ev)
		block.Event = ev
		block.Id = dev.Syspath()
		block.Timestamp = ev.Timestamp()
		if ev.Interface() != nil {
			block.Interface = ev.Interface().Name()
		}
		b, err := json.Marshal(block)
		if err != nil {
			log.Printf("unable to encode event: %v\n", b)
		}
		mu.Lock()
		os.Stdout.Write(b)
		os.Stdout.WriteString("\n")
		mu.Unlock()
	}
}

func main() {
	flag.Parse()

	monitor, err := wiimote.NewMonitor(wiimote.MonitorUdev)
	if err != nil {
		log.Fatalln("error: ", err)
	}

	fmt.Println("waiting for devices...")
	var mu sync.Mutex
	for {
		dev, err := monitor.Wait(-1)
		if err != nil || dev == nil {
			log.Printf("error while polling: %v\n", err)
			continue
		}
		go watchDevice(dev, &mu)
	}
}
