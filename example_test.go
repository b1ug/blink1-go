package blink1_test

import (
	"fmt"
	"time"

	b1 "github.com/b1ug/blink1-go"
)

// ExampleDeviceTest shows how to run test command on the blink(1) device.
func ExampleDeviceTest() {
	d, err := b1.OpenNextDevice()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	if data, err := d.Test(); err != nil {
		panic(err)
	} else {
		fmt.Println(data)
	}
}

// ExampleSetColor shows how to play a color on the blink(1) device.
func ExampleSetColor() {
	c, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	c.PlayColor(b1.ColorBlue)
}

// ExampleFadeToRGB shows how to fade to a RGB color on the blink(1) device.
func ExampleFadeToRGB() {
	c, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	st := b1.NewLightStateRGB(0, 0, 0xff, time.Second, b1.LED1)
	c.PlayState(st)
}
