package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/citilinkru/libudev"
	"github.com/citilinkru/libudev/types"
)

type JoystickEvent struct {
	Time   uint32
	Value  int16
	Type   uint8
	Number uint8
}

const (
	ButtonType uint8 = 1
	AxisType   uint8 = 1
)

func main() {
	sc := libudev.NewScanner()
	err, dvs := sc.ScanDevices()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Joysticks contain "js%d"
	rgx := regexp.MustCompile("js(\\d)+")

	filtered := []*types.Device{}

	for _, d := range dvs {
		// Find joysticks
		if !rgx.MatchString(d.Devpath) {
			continue
		}
		// Ignore mice
		if v, ok := d.Env["ID_INPUT_MOUSE"]; ok && v == "1" {
			continue
		}
		fmt.Printf("%+v\n", d)
		filtered = append(filtered, d)
	}

	for i, d := range filtered {
		fmt.Println("Filtered", i)
		go func(d *types.Device) {
			if _, ok := d.Env["DEVNAME"]; !ok {
				return
			}

			name := filepath.Join("/", "dev", d.Env["DEVNAME"])
			fmt.Println("Name", name)
			fh, err := os.Open(name)
			if err != nil {
				fmt.Println(err)
				return
			}
			for {
				e := JoystickEvent{}
				err := binary.Read(fh, binary.LittleEndian, &e)
				if err != nil {
					fmt.Println(err)
					return
				}
				// This bit determines if the event is an 'init' event,
				// which confuses me because it implies the device knows
				// which process is reading from it and is modifying its
				// data accordingly to give us different data the first
				// time we read from it but more importantly we don't
				// really care if an event is an init event or not.
				e.Type = e.Type % 128
				fmt.Printf("Device: %s, %+v\n", name, e)
			}
		}(d)
	}
	// Keep goroutines open
	for {
		runtime.Gosched()
	}
}
