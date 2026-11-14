package blink1

import (
	"errors"
	"image/color"
	"reflect"
	"testing"
	"time"

	"github.com/b1ug/colorkit"
)

func TestInvalidTypedStatesDoNotWrite(t *testing.T) {
	var typedNil *color.RGBA
	for _, tc := range []struct {
		name  string
		state LightState
		field string
	}{
		{"missing color", LightState{}, "color"},
		{"typed nil", LightState{Color: typedNil}, "color"},
		{"LED", LightState{Color: ColorRed, LED: LEDIndex(3)}, "led"},
		{"negative duration", LightState{Color: ColorRed, FadeTime: -time.Millisecond}, "fade"},
		{"sub-unit fade", LightState{Color: ColorRed, FadeTime: 15 * time.Millisecond}, "fade"},
		{"fractional ms", LightState{Color: ColorRed, FadeTime: time.Nanosecond}, "fade"},
		{"overflow", LightState{Color: ColorRed, FadeTime: 655360 * time.Millisecond}, "fade"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			controller, handle := newLifecycleController()
			defer controller.Close()
			for _, play := range []func(LightState) error{controller.PlayState, controller.PlayStateBlocking} {
				var detail *ParseError
				if err := play(tc.state); !errors.As(err, &detail) || detail.Field != tc.field {
					t.Fatalf("missing field diagnostic: %v", err)
				}
			}
			select {
			case report := <-handle.writes:
				t.Fatalf("invalid state wrote a device report: %v", report)
			default:
			}
		})
	}
}

func TestDirectRGB8StateUsesExistingControllerPath(t *testing.T) {
	controller, handle := newLifecycleController()
	defer controller.Close()
	light := colorkit.RGB8{R: 0x00, G: 0x00, B: 0xff}
	// No channel extraction or color.RGBA conversion is required by this caller.
	if err := controller.PlayState(NewLightState(light, 20*time.Millisecond, LED1)); err != nil {
		t.Fatal(err)
	}
	if got, want := readLifecycleReport(t, handle), []byte{1, 'c', 0, 0, 255, 0, 2, 1, 0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("wrong protocol report: got %v want %v", got, want)
	}
}

func TestDirectRGB8ColorKeepsGammaAndRawRGBRemainsDistinct(t *testing.T) {
	controller, handle := newLifecycleController()
	defer controller.Close()
	light := colorkit.RGB8{R: 128, G: 64, B: 32}
	if err := controller.PlayColor(light); err != nil {
		t.Fatal(err)
	}
	direct := readLifecycleReport(t, handle)
	if err := controller.PlayColor(color.RGBA{R: 128, G: 64, B: 32, A: 255}); err != nil {
		t.Fatal(err)
	}
	existing := readLifecycleReport(t, handle)
	if !reflect.DeepEqual(direct, existing) {
		t.Fatalf("direct color changed the established gamma path: %v != %v", direct, existing)
	}
	if err := controller.PlayRGB(128, 64, 32); err != nil {
		t.Fatal(err)
	}
	raw := readLifecycleReport(t, handle)
	if !reflect.DeepEqual(raw[2:5], []byte{128, 64, 32}) || reflect.DeepEqual(direct[2:5], raw[2:5]) {
		t.Fatalf("expected raw channels to remain distinct from gamma output: direct=%v raw=%v", direct, raw)
	}
	controller.SetGammaCorrection(false)
	if err := controller.PlayColor(light); err != nil {
		t.Fatal(err)
	}
	if withoutGamma := readLifecycleReport(t, handle); !reflect.DeepEqual(withoutGamma, raw) {
		t.Fatalf("disabled gamma should preserve raw channels: %v != %v", withoutGamma, raw)
	}
}

func TestNilDirectColorsDoNotWrite(t *testing.T) {
	controller, handle := newLifecycleController()
	defer controller.Close()
	var typedNil *color.RGBA
	for _, value := range []color.Color{nil, typedNil} {
		if err := controller.PlayColor(value); err == nil {
			t.Fatal("missing color was accepted")
		}
	}
	select {
	case report := <-handle.writes:
		t.Fatalf("missing color wrote to the device: %v", report)
	default:
	}
}
