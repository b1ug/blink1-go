package blink1_test

import (
	"fmt"

	b1 "github.com/b1ug/blink1-go"
)

// The public SDK parser returns LightState directly, not an intermediate DTO.
func ExampleParseStateQuery_commands() {
	for _, command := range []string{"set red", "off on led 2", "fade to blue over 2s on led 1"} {
		state, err := b1.ParseStateQuery(command)
		if err != nil {
			panic(err)
		}
		fmt.Println(state)
	}

	// Output:
	// 🎨(color=#FF0000 led=0 fade=0s)
	// 🎨(color=#000000 led=2 fade=0s)
	// 🎨(color=#0000FF led=1 fade=2s)
}

func ExampleParseStateQuery_fields() {
	state, err := b1.ParseStateQuery("led=1 color=hsv(210,75,90) fade=20ms")
	if err != nil {
		panic(err)
	}
	text, err := state.MarshalText()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(text))

	// Output:
	// #398FE6L1T20
}

// Validate the complete command before opening the hardware. Once a controller
// is open, the parsed state can be passed to PlayState without conversion.
// This hardware example is compiled, but has no Output comment and is not run.
func ExampleController_PlayState_query() {
	state, err := b1.ParseStateQuery("fade to blue over 2s on led 1")
	if err != nil {
		panic(err)
	}
	controller, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer controller.Close()
	if err := controller.PlayState(state); err != nil {
		panic(err)
	}
}
