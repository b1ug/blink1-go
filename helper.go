package blink1

import (
	"errors"
	"fmt"
	"image/color"
	"strings"
	"time"

	hid "github.com/b1ug/gid"
)

// methods in this file serve as public helper functions

// Preload triggers the initialization of the parsers and color names.
// It's optional to call this function, and it's safe to call it multiple times.
// Currently there is no noticeable performance gain from calling this function before using other APIs.
func Preload() {
	regexOnce.Do(initRegex)
	nameOnce.Do(initNames)
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
func HexToColor(hex string) (color.Color, error) {
	if len(hex) < 3 {
		return nil, errors.New("invalid hex: too short")
	}
	// remove leading #
	if strings.HasPrefix(hex, "#") {
		hex = hex[1:]
	}
	// parse
	var r, g, b uint8
	switch len(hex) {
	case 3:
		n, err := fmt.Sscanf(hex, "%1X%1X%1X", &r, &g, &b)
		if err != nil || n != 3 {
			return nil, fmt.Errorf("invalid #RGB hex: %s - %w", hex, err)
		}
		return color.RGBA{R: r * 0x11, G: g * 0x11, B: b * 0x11, A: 0xff}, nil
	case 6:
		n, err := fmt.Sscanf(hex, "%02X%02X%02X", &r, &g, &b)
		if err != nil || n != 3 {
			return nil, fmt.Errorf("invalid #RRGGBB hex: %s - %w", hex, err)
		}
		return color.RGBA{R: r, G: g, B: b, A: 0xff}, nil
	default:
		return nil, fmt.Errorf("invalid hex format: %s", hex)
	}
}

// RGBToColor converts 8-bit RGB values to color.Color.
func RGBToColor(r, g, b uint8) color.Color {
	return convRGBToColor(r, g, b)
}

// ColorToRGB converts color.Color to 8-bit RGB values.
func ColorToRGB(cl color.Color) (r, g, b uint8) {
	return convColorToRGB(cl)
}

// HexToRGB converts hex string to 8-bit RGB values.
func HexToRGB(hex string) (r, g, b uint8, err error) {
	cl, err := HexToColor(hex)
	if err != nil {
		return
	}
	r, g, b = convColorToRGB(cl)
	return
}

// RGBToHex converts 8-bit RGB values to hex string with leading #.
func RGBToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
