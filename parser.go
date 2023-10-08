package blink1

import (
	"errors"
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var colorMap = map[string]color.Color{
	"beige":   ColorBeige,
	"black":   ColorBlack,
	"blue":    ColorBlue,
	"brown":   ColorBrown,
	"cyan":    ColorCyan,
	"aqua":    ColorCyan,
	"gold":    ColorGold,
	"gray":    ColorGray,
	"grey":    ColorGray,
	"green":   ColorGreen,
	"indigo":  ColorIndigo,
	"lime":    ColorLime,
	"magenta": ColorMagenta,
	"fuchsia": ColorMagenta,
	"maroon":  ColorMaroon,
	"mint":    ColorMint,
	"navy":    ColorNavy,
	"olive":   ColorOlive,
	"orange":  ColorOrange,
	"pink":    ColorPink,
	"purple":  ColorPurple,
	"red":     ColorRed,
	"scarlet": ColorScarlet,
	"silver":  ColorSilver,
	"teal":    ColorTeal,
	"violet":  ColorViolet,
	"white":   ColorWhite,
	"yellow":  ColorYellow,
}

var (
	regexOnce         sync.Once
	colorRegexPats    = make(map[string]*regexp.Regexp)
	fadeMsecRegexPats = make(map[int]*regexp.Regexp)
	ledIdxRegexPats   = make(map[int][]*regexp.Regexp)

	errNoColorMatch = errors.New("b1: no color match")
	errNoFadeMatch  = errors.New("b1: no fade time match")
	errNoLEDMatch   = errors.New("b1: no LED index match")
	errBlankQuery   = errors.New("b1: blank query")
)

func initRegex() {
	// for colors
	colorWords := make([]string, 0, len(colorMap))
	for k := range colorMap {
		colorWords = append(colorWords, k)
	}
	colorRegexPats["name"] = regexp.MustCompile(fmt.Sprintf(`\b(%s)\b`, strings.Join(colorWords, "|")))
	colorRegexPats["on"] = regexp.MustCompile(`\b(on)\b`)
	colorRegexPats["off"] = regexp.MustCompile(`\b(off)\b`)
	colorRegexPats["rgb"] = regexp.MustCompile(`\brgb\s*\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)`)
	colorRegexPats["hsb"] = regexp.MustCompile(`\bhsb\s*\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)`)
	colorRegexPats["hex6"] = regexp.MustCompile(`#([0-9a-f]{6})\b`)
	colorRegexPats["hex3"] = regexp.MustCompile(`#([0-9a-f]{3})\b`)

	// for fade msec
	fadeMsecRegexPats[0] = regexp.MustCompile(`\b(now|immediate(?:ly)?|instant(?:ly|aneous)?(?:ly)?|quick(?:ly)?|right(?:\s)*now|swiftly|this(?:\s)*moment)\b`)
	fadeMsecRegexPats[1] = regexp.MustCompile(`\b(\d+(?:\.\d+)?)(?:\s)*(ms|millis|millisec|millisecs|msec|msecs|millisecond|milliseconds)\b`)
	fadeMsecRegexPats[1000] = regexp.MustCompile(`\b(\d+(?:\.\d+)?)(?:\s)*s(?:ec)?(?:ond)?(?:s)?\b`)
	fadeMsecRegexPats[60000] = regexp.MustCompile(`\b(\d+(?:\.\d+)?)(?:\s)*(m|min|mins|minute|minutes)\b`)

	// for led index
	ledIdxRegexPats[0] = []*regexp.Regexp{
		regexp.MustCompile(`\b(?:all\b\s*(leds|led|light|lights)?|(?:all|both)?\b\s*(?:leds|lights)|both)\b`),
	}
	ledIdxRegexPats[1] = []*regexp.Regexp{
		regexp.MustCompile(`\b(?:top|first|1st)\b\s*(led|light)\b`),
	}
	ledIdxRegexPats[2] = []*regexp.Regexp{
		regexp.MustCompile(`\b(?:bottom|second|2nd)\b\s*(led|light)\b`),
	}
	ledIdxRegexPats[12] = []*regexp.Regexp{
		regexp.MustCompile(`\b(led|light)[:#=\s]*([012])\b`),
		regexp.MustCompile(`\b(led|light)[:#=\s](all|both|zero|one|two)\b`),
	}
}

// ParseStateQuery parses the case-insensitive unstructured description of light state and returns the structured LightState.
//
// Color can be specified by name, hex code, or RGB/HSB values, e.g. "red", "#FF0000", "rgb(255,0,0)", "hsb(0,100,100)"
//
// Fade time can be specified by milliseconds, seconds, or minutes, e.g. "100ms", "1s", "1.5m", "now", "0s"
//
// LED index can be specified by number, name, or position, e.g. "led 1", "led 2", "top led", "second led", "led:all", "led:0"
func ParseStateQuery(query string) (LightState, error) {
	// init regex
	regexOnce.Do(initRegex)

	// prepare
	var st LightState
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return st, errBlankQuery
	}

	// parse each part
	if cl, err := parseColor(q); err != nil {
		return st, err
	} else {
		st.Color = cl
	}

	if ft, err := parseFadeTimeMsec(q); err != nil {
		return st, err
	} else {
		st.FadeTime = time.Duration(ft) * time.Millisecond
	}

	if led, err := parseLEDIndex(q); err != nil {
		return st, err
	} else {
		st.LED = LEDIndex(led)
	}

	// finally
	return st, nil
}

func parseColor(query string) (color.Color, error) {
	// parse
	for key, pat := range colorRegexPats {
		m := pat.FindStringSubmatch(query)

		// not match
		if m == nil {
			continue
		}

		// handle match
		val := m[1]
		switch key {
		case "name":
			return colorMap[val], nil
		case "on":
			return ColorWhite, nil
		case "off":
			return ColorBlack, nil
		case "hex6":
			var r, g, b uint8
			n, err := fmt.Sscanf(val, "%02X%02X%02X", &r, &g, &b)
			if err != nil || n != 3 {
				return nil, fmt.Errorf("invalid hex6 color: %s - %w", val, err)
			}
			return color.RGBA{R: r, G: g, B: b, A: 0xff}, nil
		case "hex3":
			var r, g, b uint8
			n, err := fmt.Sscanf(val, "%1X%1X%1X", &r, &g, &b)
			if err != nil || n != 3 {
				return nil, fmt.Errorf("invalid hex3 color: %s - %w", val, err)
			}
			return color.RGBA{R: r * 0x11, G: g * 0x11, B: b * 0x11, A: 0xff}, nil
		case "rgb":
			var r, g, b uint8
			n, err := fmt.Sscanf(m[0], "rgb(%d,%d,%d)", &r, &g, &b)
			if err != nil || n != 3 {
				return nil, fmt.Errorf("invalid rgb color: %s - %w", val, err)
			}
			return color.RGBA{R: r, G: g, B: b, A: 0xff}, nil
		case "hsb":
			var h, s, b float64
			n, err := fmt.Sscanf(m[0], "hsb(%f,%f,%f)", &h, &s, &b)
			if err != nil || n != 3 {
				return nil, fmt.Errorf("invalid hsb color: %s - %w", val, err)
			}
			return convRGBToColor(convHSBToRGB(h, s, b)), nil
		}
	}

	return nil, errNoColorMatch
}

func parseFadeTimeMsec(query string) (int, error) {
	// parse
	for mul, pat := range fadeMsecRegexPats {
		// handle values first
		if mul == 0 {
			continue
		}
		m := pat.FindStringSubmatch(query)

		// not match
		if m == nil {
			continue
		}

		// handle match
		if m != nil {
			if len(m) >= 2 {
				val, err := strconv.ParseFloat(m[1], 64)
				if err != nil {
					return 0, err
				}
				return int(val * float64(mul)), nil
			}
		}
	}

	// handle zero
	if m := fadeMsecRegexPats[0].FindStringSubmatch(query); m != nil {
		return 0, nil
	}
	return 0, errNoFadeMatch
}

func parseLEDIndex(query string) (int, error) {
	// for "led *", "light *", or "led:*" or "led#"
	for _, pat := range ledIdxRegexPats[12] {
		m := pat.FindStringSubmatch(query)
		if m != nil && len(m) >= 3 {
			switch m[2] {
			case "0", "all", "both", "zero":
				return int(0), nil
			case "1", "one":
				return int(1), nil
			case "2", "two":
				return int(2), nil
			}
		}
	}

	// for 1st led
	for _, pat := range ledIdxRegexPats[1] {
		m := pat.FindStringSubmatch(query)
		if m != nil {
			return int(1), nil
		}
	}

	// for 2nd led
	for _, pat := range ledIdxRegexPats[2] {
		m := pat.FindStringSubmatch(query)
		if m != nil {
			return int(2), nil
		}
	}

	// for all/both
	for _, pat := range ledIdxRegexPats[0] {
		m := pat.FindStringSubmatch(query)
		if m != nil {
			return int(0), nil
		}
	}

	return int(0), errNoLEDMatch
}