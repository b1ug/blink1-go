package blink1

import (
	"image/color"
	"reflect"

	"github.com/b1ug/blink1-go/textparse"
)

// Validate checks one state's executable SDK limits without accessing hardware.
// It rejects nil colors, invalid LED selectors, and fades that cannot be
// represented exactly in 10ms units up to 655350ms. Zero fade is immediate.
// PlayState calls Validate automatically. Use it separately when preflighting
// typed or decoded values before opening hardware, exporting, or queuing work.
// Errors have an empty source span because no text input was parsed.
func (state LightState) Validate() error {
	if err := validateStateColor(state.Color); err != nil {
		return err
	}
	if state.LED > LED2 {
		return &ParseError{Code: "value_out_of_range", Field: "led", Line: 1, Column: 1,
			Message: "LED must be 0, 1, or 2", Hint: "Use LEDAll, LED1, or LED2."}
	}
	return textparse.ValidateFade(state.FadeTime)
}

func validateStateColor(value color.Color) error {
	missing := value == nil
	if !missing {
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			missing = v.IsNil()
		}
	}
	if missing {
		return &ParseError{Code: "invalid_color", Field: "color", Line: 1, Column: 1,
			Message: "a non-nil color is required", Hint: "Pass a parsed color or a concrete Go color value."}
	}
	return nil
}
