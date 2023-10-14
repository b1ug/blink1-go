package blink1_test

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	b1 "github.com/b1ug/blink1-go"
)

func TestHSBToRGB(t *testing.T) {
	type hsb struct {
		hue, saturation, brightness float64
	}

	type rgb struct {
		r, g, b uint8
	}
	tests := []struct {
		name     string
		hsbValue hsb
		rgbValue rgb
	}{
		{"Black", hsb{0, 0, 0}, rgb{0, 0, 0}},
		{"White", hsb{0, 0, 100}, rgb{255, 255, 255}},
		{"Red", hsb{0, 100, 100}, rgb{255, 0, 0}},
		{"Green", hsb{120, 100, 100}, rgb{0, 255, 0}},
		{"Blue", hsb{240, 100, 100}, rgb{0, 0, 255}},
		{"Yellow", hsb{60, 100, 100}, rgb{255, 255, 0}},
		{"Cyan", hsb{180, 100, 100}, rgb{0, 255, 255}},
		{"Magenta", hsb{300, 100, 100}, rgb{255, 0, 255}},
		{"Grey", hsb{0, 0, 50}, rgb{128, 128, 128}},
		{"Violet", hsb{270, 100, 100}, rgb{128, 0, 255}},
		{"No Saturation", hsb{270, 0, 50}, rgb{128, 128, 128}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b := b1.HSBToRGB(tc.hsbValue.hue, tc.hsbValue.saturation, tc.hsbValue.brightness)
			if r != tc.rgbValue.r || g != tc.rgbValue.g || b != tc.rgbValue.b {
				t.Errorf("Expected: %v, got: R:%v, G:%v, B:%v", tc.rgbValue, r, g, b)
			}
		})
	}
}

func TestIsRunningOnSupportedOS(t *testing.T) {
	want := true
	if got := b1.IsRunningOnSupportedOS(); got != want {
		t.Errorf("IsRunningOnSupportedOS() = %v, want %v", got, want)
	}
}

func TestRandomColor(t *testing.T) {
	times := 100
	colors := make([]color.Color, times)
	for i := 0; i < times; i++ {
		colors[i] = b1.RandomColor()
	}
	counts := make(map[string]int)
	for _, c := range colors {
		r, g, b, _ := c.RGBA()
		s := fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
		counts[s]++
	}
	if lc := len(counts); lc <= int(float64(times)*0.9) {
		t.Errorf("RandomColor(*) = %v, want different colors", lc)
	}
}

func TestStringer(t *testing.T) {
	tests := []struct {
		typ fmt.Stringer
		exp string
	}{
		// as baseline
		{
			typ: time.Millisecond,
			exp: "1ms",
		},
		// for LEDIndex
		{
			typ: b1.LED1,
			exp: "LED 1",
		},
		{
			typ: b1.LED2,
			exp: "LED 2",
		},
		{
			typ: b1.LEDAll,
			exp: "All LED",
		},
		// device things
		{
			typ: b1.DevicePatternState{IsPlaying: true, CurrentPos: 1, LoopStartPos: 2, LoopEndPos: 3, RepeatTimes: 4},
			exp: "▶️{playing=true cur=1 loop=[2,3) left=4}",
		},
		{
			typ: b1.DevicePatternState{IsPlaying: false, CurrentPos: 10, LoopStartPos: 20, LoopEndPos: 30, RepeatTimes: 40},
			exp: "⏸{playing=false cur=10 loop=[20,30) left=40}",
		},
		{
			typ: b1.DeviceLightState{R: 1, G: 2, B: 3, LED: b1.LED1, FadeTimeMsec: 4},
			exp: "🎨{color=#010203 led=1 fade=4ms}",
		},
		{
			typ: b1.DeviceLightState{R: 10, G: 20, B: 30, LED: b1.LEDAll, FadeTimeMsec: 1000},
			exp: "🎨{color=#0A141E led=0 fade=1000ms}",
		},
		// controller things
		{
			typ: b1.PatternState{IsPlaying: true, CurrentPosition: 1, StartPosition: 2, EndPosition: 3, RepeatTimes: 4},
			exp: "▶️(playing=true cur=1 loop=[2,3) left=4)",
		},
		{
			typ: b1.PatternState{IsPlaying: false, CurrentPosition: 10, StartPosition: 20, EndPosition: 30, RepeatTimes: 40},
			exp: "⏸(playing=false cur=10 loop=[20,30) left=40)",
		},
		{
			typ: b1.LightState{Color: color.RGBA{R: 1, G: 2, B: 3, A: 0xff}, LED: b1.LED1, FadeTime: time.Millisecond},
			exp: "🎨(color=#010203 led=1 fade=1ms)",
		},
		{
			typ: b1.LightState{Color: color.RGBA{R: 10, G: 20, B: 30, A: 0xff}, LED: b1.LEDAll, FadeTime: time.Second},
			exp: "🎨(color=#0A141E led=0 fade=1s)",
		},
		// pattern things
		{
			typ: b1.Pattern{
				StartPosition: 0,
				EndPosition:   10,
				RepeatTimes:   1,
				States: []b1.LightState{
					{
						Color:    color.RGBA{255, 0, 0, 255},
						LED:      b1.LED1,
						FadeTime: 1 * time.Second,
					},
					{
						Color:    color.RGBA{0, 255, 0, 255},
						LED:      b1.LED2,
						FadeTime: 2 * time.Second,
					},
				},
			},
			exp: "🎼(loop=[0,10] repeat=1 states=2)",
		},
		{
			typ: b1.Pattern{
				StartPosition: 10,
				EndPosition:   20,
				RepeatTimes:   0,
			},
			exp: "🎼(loop=[10,20] repeat=∞ states=0)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.exp, func(t *testing.T) {
			if got := tc.typ.String(); got != tc.exp {
				t.Errorf("Stringer.String(%v) = %v, want %v", tc.typ, got, tc.exp)
			}
		})
	}
}
