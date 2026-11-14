package blink1_test

import (
	"fmt"

	b1 "github.com/b1ug/blink1-go"
)

func ExampleDevice_ReadPatternLineRaw() {
	// Supply an explicitly selected, already-open device. Defining this
	// function does not discover, open, or query any hardware.
	inspect := func(device *b1.Device) error {
		row, err := device.ReadPatternLineRaw(0)
		fmt.Printf("position=%d n=%d bytes=%x tail=%x\n", row.Position,
			row.Response.ReturnedLength, row.Response.Bytes(), row.Response.UnknownTail())
		if err != nil {
			return err // Keep the partial observation with its original error.
		}
		if row.RGB8 != nil && row.DurationTicks != nil && row.LEDCode != nil {
			fmt.Println(*row.RGB8, *row.DurationTicks, *row.LEDCode)
		}
		return nil
	}
	_ = inspect // No call is made by this compile-only example.
}

func ExampleDevice_ReadPlaystateRaw() {
	inspect := func(device *b1.Device) error {
		state, err := device.ReadPlaystateRaw()
		fmt.Printf("GID prefix=%x\n", state.Response.Bytes())
		if err != nil {
			return err
		}
		if state.Mode != nil && state.End != nil && state.NextPosition != nil {
			fmt.Printf("mode=%d raw_end=%d next_position=%d\n",
				*state.Mode, *state.End, *state.NextPosition)
		}
		// No bool, inclusive/exclusive conversion, or completion inference.
		return nil
	}
	_ = inspect
}
