package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/friedelschoen/go-wiimote"
	"github.com/friedelschoen/go-wiimote/driver"
	"github.com/friedelschoen/go-wiimote/pkg/discover"
	"github.com/friedelschoen/go-wiimote/pkg/eeprom"
)

func watchDevice(dev wiimote.Device) {
	fmt.Printf("new device: %s\n", dev.String())
	time.Sleep(100 * time.Millisecond)

	// coreif := wiimote.FeatureCore{}
	if err := dev.OpenFeatures(wiimote.FeatureCore, true); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open device: %s", err)
	}

	mif := dev.Feature(wiimote.FeatureCore).(wiimote.MemoryFeature)

	f, err := mif.Memory()
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

	monitor, err := discover.NewWiimoteMonitor()
	if err != nil {
		log.Fatalln("error: ", err)
	}

	fmt.Println("waiting for devices...")
	dev, err := monitor.Wait(-1)
	if err != nil || dev == nil {
		log.Printf("error while polling: %v\n", err)
		return
	}
	d, err := driver.NewDevice(dev, driver.BackendHID)
	if err != nil {
		log.Printf("error creating device: %v\n", err)
		return
	}
	watchDevice(d)
}
