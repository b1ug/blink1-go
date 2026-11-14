package blink1_test

import (
	"encoding/json"
	"errors"
	"image/color"
	"strings"
	"testing"
	"time"

	b1 "github.com/b1ug/blink1-go"
)

func TestStateSequenceEncodingPropagatesMissingColorsAtomically(t *testing.T) {
	var typedNil *color.RGBA
	valid := b1.NewLightState(b1.ColorRed, 20*time.Millisecond, b1.LED1)
	for _, bad := range []b1.LightState{{}, {Color: typedNil}} {
		for _, sequence := range []b1.StateSequence{{bad}, {valid, bad}, {bad, valid}} {
			text, err := sequence.MarshalText()
			if err == nil || text != nil || !strings.Contains(err.Error(), "state[") {
				t.Fatalf("invalid sequence was partially encoded: %q %v", text, err)
			}
			var detail *b1.ParseError
			if !errors.As(err, &detail) || detail.Field != "color" {
				t.Fatalf("lost the underlying color error: %v", err)
			}
			encoded, err := json.Marshal(sequence)
			if err == nil || encoded != nil || !errors.As(err, &detail) {
				t.Fatalf("JSON hid a sequence encoding error: %s %v", encoded, err)
			}
			encoded, err = json.Marshal(b1.Pattern{Sequence: sequence})
			if err == nil || encoded != nil {
				t.Fatalf("Pattern JSON hid a sequence encoding error: %s %v", encoded, err)
			}
		}
	}
}
