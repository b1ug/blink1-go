package blink1

import (
	"image/color"
	"time"

	hid "github.com/b1ug/gid"
)

// methods in this file serve as public helper functions

var (
	// ColorBeige is a predefined color, which is a very pale yellowish-brown color, having the RGB values #F5F5DC
	ColorBeige = color.RGBA{R: 0xF5, G: 0xF5, B: 0xDC, A: 0xFF}
	// ColorBlack is a predefined color, which absorbs all light in the visible wavelengths, having the RGB values #000000
	ColorBlack = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	// ColorBlue is a predefined color, which is one of the primary additive colors, having the RGB values #0000FF
	ColorBlue = color.RGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xFF}
	// ColorBrown is a predefined color, which is a composite color produced by mixing red, yellow, and black, having the RGB values #A52A2A
	ColorBrown = color.RGBA{R: 0xA5, G: 0x2A, B: 0x2A, A: 0xFF}
	// ColorCyan is a predefined color (a.k.a. Aqua), which is a greenish-blue color, having the RGB values #00FFFF
	ColorCyan = color.RGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0xFF}
	// ColorGold is a predefined color, which resembles the metal gold, having the RGB values #FFD700
	ColorGold = color.RGBA{R: 0xFF, G: 0xD7, B: 0x00, A: 0xFF}
	// ColorGray is a predefined color (a.k.a. Grey), which is an intermediate color between black and white, having the RGB values #808080
	ColorGray = color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xFF}
	// ColorGreen is a predefined color, which is one of the primary additive colors, having the RGB values #00FF00
	ColorGreen = color.RGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF}
	// ColorIndigo is a predefined color, which is a mix of blue and violet, having the RGB values #4B0082
	ColorIndigo = color.RGBA{R: 0x4B, G: 0x00, B: 0x82, A: 0xFF}
	// ColorLime is a predefined color, which is a fluorescent green, having the RGB values #00FF00
	ColorLime = color.RGBA{R: 0x00, G: 0x80, B: 0x00, A: 0xFF}
	// ColorMagenta is a predefined color (a.k.a. Fuchsia), which is a mix of red and blue, having the RGB values #FF00FF
	ColorMagenta = color.RGBA{R: 0xFF, G: 0x00, B: 0xFF, A: 0xFF}
	// ColorMaroon is a predefined color, which is a dark brownish red color, having the RGB values #800000
	ColorMaroon = color.RGBA{R: 0x80, G: 0x00, B: 0x00, A: 0xFF}
	// ColorMint is a predefined color, which is a pale greenish-blue color, having the RGB values #16982B
	ColorMint = color.RGBA{R: 0x16, G: 0x98, B: 0x2B, A: 0xFF}
	// ColorNavy is a predefined color, which is a very dark shade of blue, having the RGB values #000080
	ColorNavy = color.RGBA{R: 0x00, G: 0x00, B: 0x80, A: 0xFF}
	// ColorOlive is a predefined color, which resembles unripe green olives, having the RGB values #808000
	ColorOlive = color.RGBA{R: 0x80, G: 0x80, B: 0x00, A: 0xFF}
	// ColorOrange is a predefined color, which is between red and yellow, having the RGB values #FFA500
	ColorOrange = color.RGBA{R: 0xFF, G: 0xA5, B: 0x00, A: 0xFF}
	// ColorPink is a predefined color, which is a pale red color, having the RGB values #FFC0CB
	ColorPink = color.RGBA{R: 0xFF, G: 0xC0, B: 0xCB, A: 0xFF}
	// ColorPurple is a predefined color, which is a mix of red and blue, having the RGB values #800080
	ColorPurple = color.RGBA{R: 0x80, G: 0x00, B: 0x80, A: 0xFF}
	// ColorRed is a predefined color, which is one of the primary additive colors, having the RGB values #FF0000
	ColorRed = color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}
	// ColorScarlet is a predefined color, which is a bright red color with a slightly orange tinge, having the RGB values #FF2400
	ColorScarlet = color.RGBA{R: 0xFF, G: 0x24, B: 0x00, A: 0xFF}
	// ColorSilver is a predefined color, which resembles gray metallic silver, having the RGB values #C0C0C0
	ColorSilver = color.RGBA{R: 0xC0, G: 0xC0, B: 0xC0, A: 0xFF}
	// ColorTeal is a predefined color, which is a dark cyan color, having the RGB values #008080
	ColorTeal = color.RGBA{R: 0x00, G: 0x80, B: 0x80, A: 0xFF}
	// ColorWhite is a predefined color, which reflects all visible wavelengths of light, having the RGB values #FFFFFF
	ColorWhite = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	// ColorYellow is a predefined color, which is one of the primary subtractive colors, having the RGB values #FFFF00
	ColorYellow = color.RGBA{R: 0xFF, G: 0xFF, B: 0x00, A: 0xFF}
)

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

// HSBToRGB converts HSB to 8-bit RGB values.
// The hue is in degrees [0, 360], saturation and brightness/value are percent in the range [0, 100].
// Values outside of these ranges will be clamped.
func HSBToRGB(hue, saturation, brightness float64) (r, g, b uint8) {
	return convHSBToRGB(hue, saturation, brightness)
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
