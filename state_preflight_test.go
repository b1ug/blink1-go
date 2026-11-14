package blink1_test

import (
	"errors"
	"testing"
	"time"

	b1 "github.com/b1ug/blink1-go"
	"github.com/b1ug/colorkit"
)

func TestLightStatePublicPreflightNeedsNoConversion(t *testing.T) {
	state := b1.NewLightState(colorkit.RGB8{G: 255}, 20*time.Millisecond, b1.LED2)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		state b1.LightState
		field string
	}{
		{b1.LightState{}, "color"},
		{b1.NewLightState(b1.ColorRed, 15*time.Millisecond, b1.LED1), "fade"},
		{b1.NewLightState(b1.ColorRed, 0, b1.LEDIndex(255)), "led"},
	} {
		var detail *b1.ParseError
		if err := tc.state.Validate(); !errors.As(err, &detail) || detail.Field != tc.field {
			t.Fatalf("public preflight lost the field diagnostic: %v", err)
		}
	}
}
