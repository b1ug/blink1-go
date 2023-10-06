package blink1_test

import (
	"fmt"
	"image/color"
	"testing"

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
