package blink1

import "testing"

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
		{"Grey", hsb{0, 0, 50}, rgb{127, 127, 127}},
		{"No Saturation", hsb{270, 0, 50}, rgb{127, 127, 127}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b := HSBToRGB(tc.hsbValue.hue, tc.hsbValue.saturation, tc.hsbValue.brightness)
			if r != tc.rgbValue.r || g != tc.rgbValue.g || b != tc.rgbValue.b {
				t.Errorf("Expected: %v, got: R:%v, G:%v, B:%v", tc.rgbValue, r, g, b)
			}
		})
	}
}
