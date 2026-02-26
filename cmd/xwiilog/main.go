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

	"github.com/friedelschoen/go-xwiimote"
)

var (
	openIf = flag.String("interfaces", "", "interfaces to use")
)

type eventBlock struct {
	Type      string         `json:"type"`
	Event     xwiimote.Event `json:"event"`
	Id        string         `json:"id"`
	Timestamp time.Time      `json:"timestamp"`
	Interface string         `json:"interface"`
}

func watchDevice(dev *xwiimote.Device, mu *sync.Mutex) {
	fmt.Printf("new device: %s\n", dev.String())
	time.Sleep(100 * time.Millisecond)
	var ifs []xwiimote.Interface
	ifs = append(ifs, &xwiimote.InterfaceCore{})
	for name := range strings.SplitSeq(*openIf, ",") {
		switch name {
		case "accel":
			ifs = append(ifs, &xwiimote.InterfaceAccel{})
		case "bb", "balanceboard":
			ifs = append(ifs, &xwiimote.InterfaceBalanceBoard{})
		case "cc", "classiccontroller":
			ifs = append(ifs, &xwiimote.InterfaceClassicController{})
		case "drums":
			ifs = append(ifs, &xwiimote.InterfaceDrums{})
		case "guitar":
			ifs = append(ifs, &xwiimote.InterfaceGuitar{})
		case "ir":
			ifs = append(ifs, &xwiimote.InterfaceIR{})
		case "mp", "motionplus":
			ifs = append(ifs, &xwiimote.InterfaceMotionPlus{})
		case "nunchuck":
			ifs = append(ifs, &xwiimote.InterfaceNunchuck{})
		case "pc", "procontroller":
			ifs = append(ifs, &xwiimote.InterfaceProController{})
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
		if _, ok := ev.(*xwiimote.EventGone); ok {
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

	monitor, err := xwiimote.NewMonitor(xwiimote.MonitorUdev)
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
