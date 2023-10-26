package blink1

import (
	"image/color"
	"sort"
	"strings"
	"sync"
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
