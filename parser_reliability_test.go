package blink1_test

import (
	"image/color"
	"reflect"
	"strings"
	"testing"
	"time"

	b1 "github.com/b1ug/blink1-go"
)

func TestParseStateQueryFadeTimeReliability(t *testing.T) {
	const maxWholeMilliseconds = time.Duration(9223372036854) * time.Millisecond
	tests := []struct {
		query string
		want  time.Duration
	}{
		{"led 1 red 0", 0},
		{"(led:top, color:#faf, time=0)", 0},
		{"(led:top, color:#faf, time: 0)", 0},
		{"led=0 red time: 0", 0},
		{"rgb(255,0,0) led 0 0", 0},
		{"hsb(0,100,100) led 1 time=0", 0},
		{"led 0 red now", 0},
		{"led 1 red +0", 0},
		{"led 1 red 0ms", 0},
		{"led 0 red in 1s", time.Second},
		{"rgb(255,0,0) led 0 1s", time.Second},
		{"led 1 red 1s 0", time.Second},
		{"led 1 red +1s", time.Second},
		{"led 1 red. 1s", time.Second},
		{"led 1 red 1.5ms", time.Millisecond},
		{"led 1 red 0.001s", time.Millisecond},
		{"led 1 red 1s plus 1000ms", time.Second},
		{"led 1 red 1ms and 1ms", time.Millisecond},
		{"led 1 red quickly in 1 second", time.Second},
		{"led 1 red swiftly over 1 second", time.Second},
		{"led 1 red 1s // led 2 blue 2000ms", time.Second},
		{"led 1 red 9223372036854ms", maxWholeMilliseconds},
		{"led 1 red 9223372036854.775807ms", maxWholeMilliseconds},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			got, err := b1.ParseStateQuery(test.query)
			if err != nil || got.FadeTime != test.want {
				t.Fatalf("ParseStateQuery(%q): fade=%v err=%v, want %v", test.query, got.FadeTime, err, test.want)
			}
		})
	}
}

func TestParseStateQueryRejectsInvalidOrConflictingFadeTime(t *testing.T) {
	tests := []struct {
		query   string
		wantErr string
	}{
		{"led 0 red", "no fade time match"},
		{"led0 red", "no fade time match"},
		{"rgb(255,0,0) led 1", "no fade time match"},
		{"hsb(0,100,100) led 1", "no fade time match"},
		{"led 1 red 0.5", "no fade time match"},
		{"led 1 red -1s", "invalid fade time"},
		{"led 1 red -0s", "invalid fade time"},
		{"led 1 red -0.5ms", "invalid fade time"},
		{"led 1 red -0", "invalid fade time"},
		{"led 1 red 0 -0", "invalid fade time"},
		{"led 1 red 1s -0", "invalid fade time"},
		{"led 1 red 1s - 0", "invalid fade time"},
		{"led 1 red - 1s", "invalid fade time"},
		{"led 1 red --1s", "invalid fade time"},
		{"led 1 red .5s", "invalid fade time"},
		{"led 1 red 1.2.3s", "invalid fade time"},
		{"led 1 red 9223372036855ms", "invalid fade time"},
		{"led 1 red 9223372036854775808ms", "invalid fade time"},
		{"led 1 red 9223372036854.775808ms", "invalid fade time"},
		{"led 1 red " + strings.Repeat("9", 300) + "s", "invalid fade time"},
		{"led 1 red in 1s plus 2000ms", "conflicting fade times"},
		{"led 1 red in 1s plus 2s", "conflicting fade times"},
		{"led 1 red in 1s plus 2m", "conflicting fade times"},
		{"led 1 red in 0.1ms plus 0.2ms", "conflicting fade times"},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			_, err := b1.ParseStateQuery(test.query)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ParseStateQuery(%q): err=%v, want containing %q", test.query, err, test.wantErr)
			}
		})
	}
}

func TestParseStateQueryConflictingTimeIsDeterministic(t *testing.T) {
	for i := 0; i < 2048; i++ {
		_, err := b1.ParseStateQuery("led 1 red in 1s plus 2000ms")
		if err == nil || err.Error() != "b1: conflicting fade times" {
			t.Fatalf("iteration %d: expected a consistent conflict, got %v", i, err)
		}
	}
}

func TestParseColorComponentWhitespaceAndLegacySemantics(t *testing.T) {
	tests := []struct {
		query string
		want  color.RGBA
	}{
		{"rgb (1,2,3)", color.RGBA{R: 1, G: 2, B: 3, A: 255}},
		{"rgb(1,2,3 )", color.RGBA{R: 1, G: 2, B: 3, A: 255}},
		{"RGB\t( 1 ,\t2 , 3\t)", color.RGBA{R: 1, G: 2, B: 3, A: 255}},
		{"rgb(255,0,255)", color.RGBA{R: 255, B: 255, A: 255}},
		{"hsb (120,100,100)", color.RGBA{G: 255, A: 255}},
		{"hsb(120,100,100 )", color.RGBA{G: 255, A: 255}},
		{"HSB\t( 120 ,\t100 , 100\t)", color.RGBA{G: 255, A: 255}},
		{"hsb(120,101,100)", color.RGBA{G: 255, A: 255}},
		{"green", color.RGBA{G: 255, A: 255}},
		{"lime", color.RGBA{G: 128, A: 255}},
		{"#abc", color.RGBA{R: 170, G: 187, B: 204, A: 255}},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			got, err := b1.ParseColor(test.query)
			if err != nil {
				t.Fatalf("ParseColor(%q): %v", test.query, err)
			}
			if rgba := color.RGBAModel.Convert(got).(color.RGBA); rgba != test.want {
				t.Fatalf("ParseColor(%q): got %#v, want %#v", test.query, rgba, test.want)
			}
		})
	}
	if _, err := b1.ParseColor("rgb (256,0,0)"); err == nil {
		t.Fatal("out-of-range RGB component was accepted")
	}
	if repeat, err := b1.ParseRepeatTimes("repeat 0"); err != nil || repeat != 0 {
		t.Fatalf("legacy infinite repeat changed: %d, %v", repeat, err)
	}
}

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
