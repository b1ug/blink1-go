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

var (
	regexOnce         sync.Once
	titleRegexPat     *regexp.Regexp
	repeatRegexPat    *regexp.Regexp
	commentRegexPat   *regexp.Regexp
	stateTextRegexPat *regexp.Regexp
	colorRegexPats    = make(map[string]*regexp.Regexp)
	colorRegexOrder   []string
	fadeMsecRegexPats = make(map[int]*regexp.Regexp)
	ledIdxRegexPats   = make(map[int]*regexp.Regexp)

	emptyStr string

	errNoTitleMatch  = errors.New("b1: no title match")
	errNoRepeatMatch = errors.New("b1: no repeat times match")
	errNoColorMatch  = errors.New("b1: no color match")
	errNoFadeMatch   = errors.New("b1: no fade time match")
	errNoLEDMatch    = errors.New("b1: no LED index match")
	errBlankQuery    = errors.New("b1: blank query")
)

func initRegex() {
	// for simple patterns
	repeatRegexPat = regexp.MustCompile(`\brepeat\s*[:=]*\s*(\d+|\bonce|\btwice|\bthrice|\bforever|\balways|\binfinite(?:ly)?)\b|\b(infinite(?:ly)?|forever|always|once|twice|thrice)\s+repeat\b`)
	commentRegexPat = regexp.MustCompile(`(\/\/.*?$)`)
	titleRegexPat = regexp.MustCompile(`(?i)\b(title|topic|idea|subject)\s*[:=]*\s*([^\s].*?[^\s])\s*$`)
	stateTextRegexPat = regexp.MustCompile(`(?i)^#[0-9A-Fa-f]{6}L\dT\d+$`)

	// for colors
	colorWords := make([]string, 0, len(presetColorMap))
	for k := range presetColorMap {
		colorWords = append(colorWords, k)
	}
	colorRegexPats["name"] = regexp.MustCompile(fmt.Sprintf(`\b(%s)\b`, strings.Join(colorWords, "|")))
	colorRegexPats["rgb"] = regexp.MustCompile(`\brgb\s*\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)`)
	colorRegexPats["hsb"] = regexp.MustCompile(`\bhsb\s*\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)`)
	colorRegexPats["hex6"] = regexp.MustCompile(`#([0-9a-f]{6})\b`)
	colorRegexPats["hex3"] = regexp.MustCompile(`#([0-9a-f]{3})\b`)
	colorRegexPats["off"] = regexp.MustCompile(`\b(off)\b`)
	colorRegexPats["on"] = regexp.MustCompile(`\b(on)\b`)
	colorRegexOrder = []string{"name", "rgb", "hsb", "hex6", "hex3", "off", "on"}

	// for fade msec
	fadeMsecRegexPats[0] = regexp.MustCompile(`\b(0|now|immediate(?:ly)?|instant(?:ly|aneous)?(?:ly)?|quick(?:ly)?|right\s+now|swiftly|this\s+moment|no\s+time)\b`)
	fadeMsecRegexPats[1] = regexp.MustCompile(`\b(\d+(?:\.\d+)?)\s*(ms|millis|millisec|millisecs|msec|msecs|millisecond|milliseconds)\b`)
	fadeMsecRegexPats[1000] = regexp.MustCompile(`\b(\d+(?:\.\d+)?)\s*s(?:ec)?(?:ond)?(?:s)?\b`)
	fadeMsecRegexPats[60000] = regexp.MustCompile(`\b(\d+(?:\.\d+)?)\s*(m|min|mins|minute|minutes)\b`)

	// for led index
	ledIdxRegexPats[0] = regexp.MustCompile(`\b(?:all\s+(leds|led|light|lights)?|(?:all|both)?\s+(?:leds|lights)|both)\b`)
	ledIdxRegexPats[1] = regexp.MustCompile(`\b(?:top|first|1st)\s+(led|light)\b`)
	ledIdxRegexPats[2] = regexp.MustCompile(`\b(?:btm|bottom|second|2nd)\s+(led|light)\b`)
	ledIdxRegexPats[12] = regexp.MustCompile(`\b(led|light)[:#=\s]*([012]|top|bottom|btm|all|both|zero|one|two)\b`)
}

// ParseTitle parses the labeled title or topic or idea string from the query string. It returns the title or an error if no title is found.
func ParseTitle(query string) (string, error) {
	// init regex
	regexOnce.Do(initRegex)

	// match
	q := strings.TrimSpace(query)
	m := titleRegexPat.FindStringSubmatch(q)
	if len(m) <= 2 {
		// We now need match of length > 2 as our pattern has a second capture group
		return emptyStr, errNoTitleMatch
	}

	// handle match
	title := m[2]
	if title == emptyStr {
		return emptyStr, errNoTitleMatch
	}
	return title, nil
}

// ParseRepeatTimes parses the case-insensitive unstructured description of repeat times and returns the number of times to repeat.
func ParseRepeatTimes(query string) (uint, error) {
	// init regex
	regexOnce.Do(initRegex)

	// match
	q := strings.TrimSpace(strings.ToLower(query))
	m := repeatRegexPat.FindStringSubmatch(q)
	if len(m) <= 1 {
		return 0, errNoRepeatMatch
	}

	// handle match
	var r string
	for i := 1; i < len(m); i++ {
		if m[i] != emptyStr {
			r = m[i]
			break
		}
	}
	switch r {
	case "0", "forever", "always", "infinite", "infinitely":
		return 0, nil
	case "once":
		return 1, nil
	case "twice":
		return 2, nil
	case "thrice":
		return 3, nil
	case emptyStr:
		return 0, errNoRepeatMatch
	default:
		times, err := strconv.Atoi(r)
		if err != nil {
			return 0, fmt.Errorf("b1: conversion error: %w", err)
		}
		return uint(times), nil
	}
}

// ParseColor parses the case-insensitive unstructured description of color and returns the corresponding color.Color.
func ParseColor(query string) (color.Color, error) {
	// init regex
	regexOnce.Do(initRegex)

	// prepare
	query = strings.TrimSpace(strings.ToLower(query))
	if query == emptyStr {
		return nil, errBlankQuery
	}

	// parse
	return parseColorQuery(query)
}

// ParseStateQuery parses the case-insensitive unstructured description of light state and returns the structured LightState.
// The query can contain information about the color, fade time, and LED index. For example, "turn off all lights right now", "set led 1 to color #ff00ff over 2 sec", "#FF0000L1T500".
// If the query is empty, it returns an error.
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
	var state LightState
	query = strings.TrimSpace(strings.ToLower(query))
	if query == emptyStr {
		return state, errBlankQuery
	}

	// remove comments
	query = commentRegexPat.ReplaceAllString(query, emptyStr)

	// attempt to parse full as state
	if stateTextRegexPat.MatchString(query) {
		var st LightState
		if err := st.UnmarshalText([]byte(query)); err == nil {
			return st, nil
		}
	}

	// parse each part
	var err error
	if state.Color, err = parseColorQuery(query); err != nil {
		return state, err
	}
	if state.FadeTime, err = parseFadeTime(query); err != nil {
		return state, err
	}
	if state.LED, err = parseLEDIndex(query); err != nil {
		return state, err
	}

	// all done
	return state, nil
}

func parseColorQuery(query string) (color.Color, error) {
	// parse
	for _, key := range colorRegexOrder {
		pat, ok := colorRegexPats[key]
		if !ok {
			continue
		}

		m := pat.FindStringSubmatch(query)
		if m == nil {
			// not match
			continue
		}

		// handle match
		val := m[1]
		switch key {
		case "name":
			return presetColorMap[val], nil
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

func parseFadeTime(query string) (time.Duration, error) {
	// parse
	for mul, pat := range fadeMsecRegexPats {
		// skip zero values first
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
				return time.Duration(val*float64(mul)) * time.Millisecond, nil
			}
		}
	}

	// handle zero values
	if m := fadeMsecRegexPats[0].FindStringSubmatch(query); m != nil {
		return 0, nil
	}
	return 0, errNoFadeMatch
}

func parseLEDIndex(query string) (LEDIndex, error) {
	// for "led *", "light *", or "led:*" or "led#"
	if m := ledIdxRegexPats[12].FindStringSubmatch(query); m != nil && len(m) >= 3 {
		switch m[2] {
		case "0", "all", "both", "zero":
			return LEDAll, nil
		case "1", "one", "top":
			return LED1, nil
		case "2", "two", "btm", "bottom":
			return LED2, nil
		}
	}

	// for 1st led
	if m := ledIdxRegexPats[1].FindStringSubmatch(query); m != nil {
		return LED1, nil
	}

	// for 2nd led
	if m := ledIdxRegexPats[2].FindStringSubmatch(query); m != nil {
		return LED2, nil
	}

	// for all/both
	if m := ledIdxRegexPats[0].FindStringSubmatch(query); m != nil {
		return LEDAll, nil
	}

	// no match
	return 0, errNoLEDMatch
}
