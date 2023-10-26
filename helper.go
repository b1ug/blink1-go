package blink1

import (
	"image/color"
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
