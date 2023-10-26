package blink1

import (
	"image/color"
	"strings"
)

// GetColorNames returns the color names from the preset color map.
func GetColorNames() []string {
	// copy name slice
	cls := make([]string, len(colorNames))
	copy(cls, colorNames)
	return cls
}

// GetColorByName returns the color corresponding to the given name from the preset color map.
// If the color is found, it returns the color and true, otherwise it returns nil and false.
func GetColorByName(name string) (cl color.Color, found bool) {
	n := strings.TrimSpace(strings.ToLower(name))
	cl, found = presetColorMap[n]
	return
}

// GetNameByColor returns the name corresponding to the given color from the preset color map.
// If the color is found, it returns the name and true.
// If the color is not found, it returns the hex string and false.
func GetNameByColor(cl color.Color) (name string, found bool) {
	// check if color is in map
	if name, ok := hexNameMap[convColorToHex(cl)]; ok {
		return name, true
	}
	return convColorToHex(cl), false
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
