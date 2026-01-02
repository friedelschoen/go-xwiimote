package main

import (
	"fmt"
	"log"

	"github.com/friedelschoen/go-xwiimote/pkg/udev"
)

func main() {
	e := udev.NewEnumerate()

	if err := e.AddMatchSubsystem("input"); err != nil {
		log.Fatalln("error: ", err)
	}
	if err := e.AddMatchSubsystem("leds"); err != nil {
		log.Fatalln("error: ", err)
	}
	if err := e.AddMatchSubsystem("power_supply"); err != nil {
		log.Fatalln("error: ", err)
	}
	dev := udev.NewDeviceFromSyspath("/sys/devices/virtual/misc/uhid/0005:057E:0306.0005")
	e.AddMatchParent(dev)
	iter, err := e.Devices()
	if err != nil {
		log.Fatalln("error: ", err)
	}
	for dev := range iter {
		if err != nil {
			log.Println("warn: ", err)
			continue
		}
		fmt.Println("dev: ", dev)
	}
}
