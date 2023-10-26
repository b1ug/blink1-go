package blink1_test

import (
	"fmt"
	"time"

	b1 "github.com/b1ug/blink1-go"
)

// This example checks if the current OS is supported by the underlying HID library.
func ExampleIsRunningOnSupportedOS() {
	fmt.Println(b1.IsRunningOnSupportedOS())
}

// This example shows how to run test command on the blink(1) device.
func ExampleDevice_Test() {
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

// This example shows how to play a color on the blink(1) device.
func ExampleController_PlayColor() {
	c, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	c.PlayColor(b1.ColorBlue)
}

// This example shows how to fade to a RGB color on the blink(1) device.
func ExampleController_PlayState() {
	c, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	st := b1.NewLightStateRGB(0, 0, 0xff, time.Second, b1.LED1)
	c.PlayState(st)
}

// This example shows how to read the current color of the first LED of the blink(1) device.
func ExampleController_ReadColor() {
	c, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	if cl, err := c.ReadColor(b1.LED1); err != nil {
		panic(err)
	} else {
		fmt.Println(cl)
	}
}

// This example illustrates how to generate a state sequence for Hawaiian rainbow on the blink(1) device, looping it thrice, and pausing until the execution completes.
func ExampleController_PlayPatternBlocking() {
	// build a rainbow sequence with 2 states for each color, one for fade in and one for maintain
	seq := make(b1.StateSequence, len(b1.RainbowColors)*2)
	for i, cl := range b1.RainbowColors {
		st := b1.NewLightState(cl, 300*time.Millisecond, b1.LEDAll)
		seq[i*2] = st
		seq[i*2+1] = st
	}
	pat := b1.Pattern{
		StartPosition: 0,
		EndPosition:   uint(seq.Length() - 1),
		RepeatTimes:   3,
		Sequence:      seq,
	}

	// open the device and play the pattern
	c, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	c.PlayPatternBlocking(pat)

	// turn off all LEDs
	c.StopPlaying()
}

// This example shows how to get a random color.
func ExampleRandomColor() {
	cl := b1.RandomColor()
	fmt.Println(cl)
}

// This example shows how to parse a title query.
func ExampleParseTitle() {
	t, err := b1.ParseTitle("title: Hawaiian Rainbow")
	if err != nil {
		panic(err)
	}
	fmt.Println(t)

	// Output:
	// Hawaiian Rainbow
}

// This example shows how to parse a repeat times query.
func ExampleParseRepeatTimes() {
	rt, err := b1.ParseRepeatTimes("Repeat 3 times")
	if err != nil {
		panic(err)
	}
	fmt.Println(rt)

	// Output:
	// 3
}

// This example shows how to parse a state query.
func ExampleParseStateQuery() {
	st, err := b1.ParseStateQuery("Fade the first LED to blue in 1.5 seconds.")
	if err != nil {
		panic(err)
	}
	fmt.Println(st)

	// Output:
	// 🎨(color=#0000FF led=1 fade=1.5s)
}
