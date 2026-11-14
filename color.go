package blink1

import (
	"image/color"
	"strings"

	"github.com/b1ug/colorkit"
)

// GetColorNames returns a sorted copy of the 148 supported opaque CSS names.
func GetColorNames() []string {
	return colorkit.Names()
}

// GetColorByName looks up a CSS name using colorkit's bounded ASCII rules.
// A successful result remains a concrete color.RGBA value with full opacity.
// If the color is found, it returns the color and true, otherwise it returns nil and false.
func GetColorByName(name string) (cl color.Color, found bool) {
	rgb, ok := colorkit.LookupName(name)
	if !ok {
		return nil, false
	}
	return color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 0xff}, true
}

// cssHexNames is a derived reverse index, not a second named-color table.
var cssHexNames = makeCSSHexNames()

func makeCSSHexNames() map[string]string {
	names := make(map[string]string)
	// Names is sorted: otherwise equivalent aliases have a stable first choice.
	for _, name := range colorkit.Names() {
		rgb, _ := colorkit.LookupName(name)
		hex := strings.ToUpper(rgb.Hex())
		if _, exists := names[hex]; !exists {
			names[hex] = name
		}
	}
	// Preserve established SDK preferences and use gray rather than grey.
	for _, name := range []string{"cyan", "magenta", "gray", "darkgray", "darkslategray", "dimgray", "lightgray", "lightslategray", "slategray"} {
		rgb, _ := colorkit.LookupName(name)
		names[strings.ToUpper(rgb.Hex())] = name
	}
	return names
}

// GetNameByColor returns a deterministic CSS name for the given color.
// It prefers cyan, magenta, and gray spellings when aliases share a value.
// If the color is found, it returns the name and true.
// Otherwise it returns uppercase Hex and false. Existing alpha-ignoring,
// high-eight-bit conversion behavior is retained.
func GetNameByColor(cl color.Color) (name string, found bool) {
	hex := convColorToHex(cl)
	if name, ok := cssHexNames[hex]; ok {
		return name, true
	}
	return hex, false
}

// GetNameOrHexByColor returns the name corresponding to the given color from the preset color map, or the hex string if the color is not found.
func GetNameOrHexByColor(cl color.Color) string {
	name, _ := GetNameByColor(cl)
	return name
}

// RandomColor returns a bright random color.
func RandomColor() color.Color {
	// helper function to get a random float64
	rand := func(mul float64) float64 {
		f, _ := getRandomFloat(1 << 16)
		return f * mul
	}
	// hue between 0 and 360 to get a full range of colors
	hue := rand(360)
	// saturation between 50 and 100 to ensure a bright color
	saturation := 50. + rand(50)
	// max brightness for a bright color
	brightness := 90. + rand(10)
	// convert to RGB and return
	return convRGBToColor(convHSBToRGB(hue, saturation, brightness))
}

// DecodeGammaColor operates gamma correction on the given color and returns the decoded color.
// The gamma correction is automatically applied for state and pattern while playing or writing.
// This is the inverse function of EncodeGammaColor.
func DecodeGammaColor(cl color.Color) color.Color {
	return convRGBToColor(degammaRGB(convColorToRGB(cl)))
}

// EncodeGammaColor operates gamma correction on the given color and returns the encoded color.
// Encoding a decoded color may not return the original color, but it will be the same color when decoded again.
func EncodeGammaColor(cl color.Color) color.Color {
	return convRGBToColor(engammaRGB(convColorToRGB(cl)))
}
