package blink1

import (
	"image/color"
	"strings"
	"time"

	"github.com/b1ug/colorkit"
	hid "github.com/b1ug/gid"
)

// methods in this file serve as public helper functions

// Preload initializes the metadata-helper and compact-state regular expressions.
// Complete color and command parsing does not require preloading.
// It's optional to call this function, and it's safe to call it multiple times.
// Currently there is no noticeable performance gain from calling this function before using other APIs.
func Preload() {
	regexOnce.Do(initRegex)
}

// IsRunningOnSupportedOS returns true if the current OS is supported by underlying HID library.
func IsRunningOnSupportedOS() bool {
	return hid.Supported()
}

// IsBlink1Device returns true if the device info is about a blink(1) device.
func IsBlink1Device(di *hid.DeviceInfo) bool {
	if di == nil {
		return false
	}
	if di.VendorID == b1VendorID && di.ProductID == b1ProductID {
		return true
	}
	return false
}

// NewLightState returns a new LightState with the given color and fade time.
func NewLightState(cl color.Color, fadeTime time.Duration, ledN LEDIndex) LightState {
	return LightState{
		Color:    cl,
		LED:      ledN,
		FadeTime: fadeTime,
	}
}

// NewLightStateRGB returns a new LightState with the given RGB color and fade time.
func NewLightStateRGB(r, g, b uint8, fadeTime time.Duration, ledN LEDIndex) LightState {
	return LightState{
		Color:    convRGBToColor(r, g, b),
		LED:      ledN,
		FadeTime: fadeTime,
	}
}

// NewLightStateHSB returns a new LightState with the given HSB/HSV color and fade time.
// Valid hue range is [0, 360], saturation range and brightness/value range is [0, 100].
func NewLightStateHSB(h, s, b float64, fadeTime time.Duration, ledN LEDIndex) LightState {
	return LightState{
		Color:    convHSBToColor(h, s, b),
		LED:      ledN,
		FadeTime: fadeTime,
	}
}

// HSBToRGB converts HSB to 8-bit RGB values.
// e.g. 0, 100, 100 -> 0xff, 0x00, 0x00
// The hue is in degrees [0, 360], saturation and brightness/value are percent in the range [0, 100].
// Values outside of these ranges will be clamped.
func HSBToRGB(hue, sat, bright float64) (red, green, blue uint8) {
	return convHSBToRGB(hue, sat, bright)
}

// ColorToHex converts color.Color to hex string with leading #.
// e.g. color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff} -> "#FF0000"
// However, if you need color names instead of hex strings, use GetNameByColor()/GetNameOrHexByColor() methods instead.
func ColorToHex(cl color.Color) string {
	return convColorToHex(cl)
}

// HexToColor converts hex string to color.Color. The hex string can be in the format of #RRGGBB or #RGB or RRGGBB or RGB (case insensitive).
// It accepts surrounding ASCII whitespace and requires the complete input to be
// one hexadecimal literal. The original input is limited to 64 ASCII bytes.
// Success returns color.RGBA; errors support errors.Is/As with colorkit errors.
// e.g. "#FF0000" -> color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
// However, if you want to convert color names to color.Color, use GetColorByName() instead.
// And if your input string is more complex and diverse than hex or names, use ParseColor() instead.
func HexToColor(hex string) (color.Color, error) {
	rgb, err := parseHexColor(hex)
	if err != nil {
		return nil, err
	}
	return convRGBToColor(rgb.R, rgb.G, rgb.B), nil
}

// parseHexColor applies the SDK's optional-prefix grammar once. Keeping the
// RGB8 result avoids converting through color.RGBA and 16-bit channels when
// a caller requested three RGB bytes directly.
func parseHexColor(hex string) (colorkit.RGB8, error) {
	literal := hex
	// Add the optional prefix only after respecting the original input bound.
	// Oversized input reaches the core unchanged so it returns its typed error.
	if len(hex) <= 64 {
		trimmed := strings.Trim(hex, " \t\r\n\v\f")
		if !strings.HasPrefix(trimmed, "#") {
			literal = "#" + trimmed
		}
	}
	return colorkit.ParseHex(literal)
}

// RGBToColor converts 8-bit RGB values to color.Color.
// e.g. 0xff, 0x00, 0x00 -> color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
func RGBToColor(r, g, b uint8) color.Color {
	return convRGBToColor(r, g, b)
}

// ColorToRGB converts color.Color to 8-bit RGB values.
// e.g. color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff} -> 0xff, 0x00, 0x00
func ColorToRGB(cl color.Color) (r, g, b uint8) {
	return convColorToRGB(cl)
}

// HexToRGB converts hex string to 8-bit RGB values. It shares the same complete-literal validation as HexToColor().
// e.g. "#FF0000" -> 0xff, 0x00, 0x00
func HexToRGB(hex string) (r, g, b uint8, err error) {
	rgb, err := parseHexColor(hex)
	if err != nil {
		return 0, 0, 0, err
	}
	return rgb.R, rgb.G, rgb.B, nil
}

// RGBToHex converts 8-bit RGB values to hex string with leading #.
// e.g. 0xff, 0x00, 0x00 -> "#FF0000"
func RGBToHex(r, g, b uint8) string {
	return strings.ToUpper(colorkit.RGB8{R: r, G: g, B: b}.Hex())
}
