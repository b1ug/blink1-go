package blink1_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	b1 "github.com/b1ug/blink1-go"
)

func TestParseTitleSingleCharacter(t *testing.T) {
	for _, title := range []string{"A", "灯", "💡"} {
		for _, prefix := range []string{"title: ", "topic=", "idea:::", "subject "} {
			query := prefix + title
			got, err := b1.ParseTitle(query)
			if err != nil || got != title {
				t.Fatalf("ParseTitle(%q): got %q err=%v, want %q", query, got, err, title)
			}
		}
	}
	for _, query := range []string{"title", "title:", "topic= ", "idea::: ", "subject  "} {
		if got, err := b1.ParseTitle(query); err == nil {
			t.Errorf("ParseTitle(%q): accepted empty title %q", query, got)
		}
	}
}

func TestLightStateUnmarshalTextBoundaries(t *testing.T) {
	const maxWholeMilliseconds = time.Duration(9223372036854) * time.Millisecond
	valid := []struct {
		text string
		led  b1.LEDIndex
		fade time.Duration
	}{
		{"#ff0000l0t0", b1.LEDAll, 0},
		{"#ff0000L1T1", b1.LED1, time.Millisecond},
		{"#ff0000L2T9223372036854", b1.LED2, maxWholeMilliseconds},
		{"#ff0000L01T001", b1.LED1, time.Millisecond},
	}
	for _, test := range valid {
		var got b1.LightState
		if err := got.UnmarshalText([]byte(test.text)); err != nil || got.LED != test.led || got.FadeTime != test.fade {
			t.Errorf("UnmarshalText(%q): got=%v err=%v, want led=%v fade=%v", test.text, got, err, test.led, test.fade)
		}
	}
	invalid := []string{
		"", "#ff0000L3T1", "#ff0000L9T1", "#ff0000L255T1", "#ff0000L-1T1",
		"#ff0000L9999999999999999999999T1", "#ff0000L1T-1", "#ff0000L1T-0",
		"#ff0000L1T9223372036855", "#ff0000L1T9223372036854775807",
		"#ff0000L1T9223372036854775808", "#ff0000L1T1garbage", "#ff0000L1T1;",
		"#ff0000M1T1", "#gg0000L1T1", "#ff0000 L1T1",
	}
	before := b1.LightState{Color: b1.ColorBlue, LED: b1.LED2, FadeTime: 123 * time.Millisecond}
	for _, text := range invalid {
		t.Run(text, func(t *testing.T) {
			got := before
			if err := got.UnmarshalText([]byte(text)); err == nil {
				t.Fatalf("UnmarshalText(%q): accepted invalid state %v", text, got)
			}
			if !reflect.DeepEqual(got, before) {
				t.Fatalf("UnmarshalText(%q): invalid input changed receiver from %v to %v", text, before, got)
			}
		})
	}
}

func TestParseStateQueryPreservesCompactValidationError(t *testing.T) {
	for _, test := range []struct {
		query   string
		wantErr string
	}{
		{"#ff0000L9T500", "invalid LED index"},
		{"#ff0000L1T-1", "invalid fade time"},
		{"#ff0000L1T9223372036854775807", "invalid fade time"},
	} {
		_, err := b1.ParseStateQuery(test.query)
		if err == nil || !strings.Contains(err.Error(), test.wantErr) {
			t.Errorf("ParseStateQuery(%q): err=%v, want containing %q", test.query, err, test.wantErr)
		}
	}
}

func TestStateSequenceInvalidEntryDoesNotChangeReceiver(t *testing.T) {
	before := b1.StateSequence{{Color: b1.ColorBlue, LED: b1.LED2, FadeTime: 123 * time.Millisecond}}
	got := append(b1.StateSequence(nil), before...)
	err := got.UnmarshalText([]byte("#ff0000L1T5;#ff0000L9T5"))
	if err == nil || !reflect.DeepEqual(got, before) {
		t.Fatalf("invalid sequence changed receiver: got=%v err=%v, want %v", got, err, before)
	}
}
