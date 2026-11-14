package blink1_test

import (
	"fmt"
	"time"

	b1 "github.com/b1ug/blink1-go"
	"github.com/b1ug/colorkit"
)

// colorkit values already implement color.Color; pass them to the SDK directly.
func ExampleNewLightState_colorkit() {
	light, err := colorkit.Parse("#abc")
	if err != nil {
		panic(err)
	}
	state := b1.NewLightState(light, 250*time.Millisecond, b1.LED1)
	fmt.Println(state)

	// Output:
	// 🎨(color=#AABBCC led=1 fade=250ms)
}

// HSV validates and converts once, then the resulting value goes straight into
// the SDK. H is in degrees; S and V are percentages.
func ExampleNewLightState_hsv() {
	light, err := (colorkit.HSV{H: 210, S: 75, V: 90}).RGB8()
	if err != nil {
		panic(err)
	}
	state := b1.NewLightState(light, time.Second, b1.LEDAll)
	fmt.Println(state)

	// Output:
	// 🎨(color=#398FE6 led=0 fade=1s)
}

// The existing SDK helper is a single-call option for hexadecimal input.
func ExampleHexToColor_lightState() {
	light, err := b1.HexToColor("abc")
	if err != nil {
		panic(err)
	}
	state := b1.NewLightState(light, 500*time.Millisecond, b1.LED2)
	fmt.Println(state)

	// Output:
	// 🎨(color=#AABBCC led=2 fade=500ms)
}

// This example requires connected hardware. It is compiled, but has no Output
// comment, so go test does not open a controller or execute PlayColor.
func ExampleController_PlayColor_colorkit() {
	light, err := colorkit.Parse("rebeccapurple")
	if err != nil {
		panic(err)
	}
	controller, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer controller.Close()

	if err := controller.PlayColor(light); err != nil {
		panic(err)
	}
}
