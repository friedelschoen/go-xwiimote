package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/pkg/eeprom"
)

func watchDevice(dev *wiimote.Device) {
	fmt.Printf("new device: %s\n", dev.String())
	time.Sleep(100 * time.Millisecond)
	coreif := wiimote.InterfaceCore{}
	if err := dev.OpenInterfaces(true, &coreif); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open device: %s", err)
	}

	f, err := coreif.Memory()
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	block, err := eeprom.ReadMiiBlock(f)
	if err != nil {
		log.Fatalln(err)
	}
	for slot := range block.MiiSlotSeq() {
		mii := eeprom.DecodeMii(slot)
		fmt.Printf("%q by %q\n", mii.Name, mii.CreatorName)
	}
}

func main() {
	flag.Parse()

	monitor, err := wiimote.NewMonitor(wiimote.MonitorUdev)
	if err != nil {
		log.Fatalln("error: ", err)
	}

	fmt.Println("waiting for devices...")
	dev, err := monitor.Wait(-1)
	if err != nil || dev == nil {
		log.Printf("error while polling: %v\n", err)
		return
	}
	watchDevice(dev)
}
