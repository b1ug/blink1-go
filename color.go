package blink1

import (
	"image/color"
	"sort"
	"strings"
	"sync"
)

var (
	// ColorApricot is a predefined color, it's a light orange color similar to the color of an apricot fruit, having the RGB values #FBCEB1
	ColorApricot = color.RGBA{R: 0xFB, G: 0xCE, B: 0xB1, A: 0xFF}
	// ColorBeige is a predefined color, which is a very pale yellowish-brown color, having the RGB values #F5F5DC
	ColorBeige = color.RGBA{R: 0xF5, G: 0xF5, B: 0xDC, A: 0xFF}
	// ColorBlack is a predefined color, which absorbs all light in the visible wavelengths, having the RGB values #000000
	ColorBlack = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	// ColorBlue is a predefined color, which is one of the primary additive colors, having the RGB values #0000FF
	ColorBlue = color.RGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xFF}
	// ColorBronze is a predefined color, it resembles the color of the metal bronze, having the RGB values #CD7F32
	ColorBronze = color.RGBA{R: 0xCD, G: 0x7F, B: 0x32, A: 0xFF}
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
	// ColorLavender is a predefined color, it's a light purple color similar to lavender flowers, having the RGB values #E6E6FA
	ColorLavender = color.RGBA{R: 0xE6, G: 0xE6, B: 0xFA, A: 0xFF}
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
	// ColorPeach is a predefined color, it's similar to the color of a peach fruit, having the RGB values #FFE5B4
	ColorPeach = color.RGBA{R: 0xFF, G: 0xE5, B: 0xB4, A: 0xFF}
	// ColorPink is a predefined color, which is a pale red color, having the RGB values #FFC0CB
	ColorPink = color.RGBA{R: 0xFF, G: 0xC0, B: 0xCB, A: 0xFF}
	// ColorPlum is a predefined color, it's a dark purple color similar to the color of a plum fruit, having the RGB values #8E4585
	ColorPlum = color.RGBA{R: 0x8E, G: 0x45, B: 0x85, A: 0xFF}
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
	// ColorViolet is a predefined color, which is a mix of blue and red, having the RGB values #8000FF
	ColorViolet = color.RGBA{R: 0x80, G: 0x00, B: 0xFF, A: 0xFF}
	// ColorWhite is a predefined color, which reflects all visible wavelengths of light, having the RGB values #FFFFFF
	ColorWhite = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	// ColorYellow is a predefined color, which is one of the primary subtractive colors, having the RGB values #FFFF00
	ColorYellow = color.RGBA{R: 0xFF, G: 0xFF, B: 0x00, A: 0xFF}

	// RainbowColors is a predefined color palette, which contains the 7 colors of the rainbow.
	RainbowColors = []color.Color{ColorRed, ColorOrange, ColorYellow, ColorGreen, ColorCyan, ColorBlue, ColorViolet}
)

// presetColorMap is a map of all supported preset color names to color values.
var presetColorMap = map[string]color.Color{
	"apricot":  ColorApricot,
	"aqua":     ColorCyan,
	"beige":    ColorBeige,
	"black":    ColorBlack,
	"blue":     ColorBlue,
	"bronze":   ColorBronze,
	"brown":    ColorBrown,
	"cyan":     ColorCyan,
	"fuchsia":  ColorMagenta,
	"gold":     ColorGold,
	"gray":     ColorGray,
	"green":    ColorGreen,
	"grey":     ColorGray,
	"indigo":   ColorIndigo,
	"lavender": ColorLavender,
	"lime":     ColorLime,
	"magenta":  ColorMagenta,
	"maroon":   ColorMaroon,
	"mint":     ColorMint,
	"navy":     ColorNavy,
	"olive":    ColorOlive,
	"orange":   ColorOrange,
	"peach":    ColorPeach,
	"pink":     ColorPink,
	"plum":     ColorPlum,
	"purple":   ColorPurple,
	"red":      ColorRed,
	"scarlet":  ColorScarlet,
	"silver":   ColorSilver,
	"teal":     ColorTeal,
	"violet":   ColorViolet,
	"white":    ColorWhite,
	"yellow":   ColorYellow,
}

var (
	nameOnce   sync.Once
	colorNames []string
	hexNameMap map[string]string
	emptyStr   string
)

func initNames() {
	colorNames = make([]string, 0, len(presetColorMap))
	hexNameMap = make(map[string]string, len(presetColorMap))
	for name, col := range presetColorMap {
		colorNames = append(colorNames, name)
		hexNameMap[convColorToHex(col)] = name
	}
	sort.Strings(colorNames)
}

// GetColorNames returns the color names from the preset color map.
func GetColorNames() []string {
	// init name maps
	nameOnce.Do(initNames)
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
	// init name maps
	nameOnce.Do(initNames)
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

// HSBToRGB converts HSB to 8-bit RGB values.
// The hue is in degrees [0, 360], saturation and brightness/value are percent in the range [0, 100].
// Values outside of these ranges will be clamped.
func HSBToRGB(hue, sat, bright float64) (red, green, blue uint8) {
	return convHSBToRGB(hue, sat, bright)
}
