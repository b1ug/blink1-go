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
	fadeTimeRegexPats []fadeTimePattern
	immediateRegexPat *regexp.Regexp
	zeroTimeRegexPat  *regexp.Regexp
	ledIdxRegexPats   = make(map[int]*regexp.Regexp)

	emptyStr string

	errNoTitleMatch  = errors.New("b1: no title match")
	errNoRepeatMatch = errors.New("b1: no repeat times match")
	errNoColorMatch  = errors.New("b1: no color match")
	errNoFadeMatch   = errors.New("b1: no fade time match")
	errNoLEDMatch    = errors.New("b1: no LED index match")
	errBlankQuery    = errors.New("b1: blank query")
	errAmbiguousFade = errors.New("b1: conflicting fade times")
)

type fadeTimePattern struct {
	unit string
	pat  *regexp.Regexp
}

func initRegex() {
	// for simple patterns
	repeatRegexPat = regexp.MustCompile(`\brepeat\s*[:=]*\s*(\d+|\bonce|\btwice|\bthrice|\bforever|\balways|\binfinite(?:ly)?)\b|\b(infinite(?:ly)?|forever|always|once|twice|thrice)\s+repeat\b`)
	commentRegexPat = regexp.MustCompile(`(\/\/.*?$)`)
	titleRegexPat = regexp.MustCompile(`(?i)\b(title|topic|idea|subject)\s*[:=]*\s*`)
	stateTextRegexPat = regexp.MustCompile(`(?i)^#([0-9a-f]{6})L([+-]?\d+)T([+-]?\d+)$`)

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
	immediateRegexPat = regexp.MustCompile(`\b(now|immediate(?:ly)?|instant(?:ly|aneous)?(?:ly)?|quick(?:ly)?|right\s+now|swiftly|this\s+moment|no\s+time)\b`)
	zeroTimeRegexPat = regexp.MustCompile(`[+-]?\b0\b`)
	fadeTimeRegexPats = []fadeTimePattern{
		{"ms", regexp.MustCompile(`([+-]?\b\d+(?:\.\d+)?)\s*(ms|millis|millisec|millisecs|msec|msecs|millisecond|milliseconds)\b`)},
		{"s", regexp.MustCompile(`([+-]?\b\d+(?:\.\d+)?)\s*s(?:ec)?(?:ond)?(?:s)?\b`)},
		{"m", regexp.MustCompile(`([+-]?\b\d+(?:\.\d+)?)\s*(m|min|mins|minute|minutes)\b`)},
	}

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
	m := titleRegexPat.FindStringIndex(q)
	if m == nil {
		return emptyStr, errNoTitleMatch
	}

	// handle match
	title := strings.TrimSpace(q[m[1]:])
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
// This is a field extractor, not a general natural-language command interpreter:
// it does not interpret action verbs, negation, or multiple instructions.
//
// Color can be specified by name, hex code, or RGB/HSB values, e.g. "red", "#FF0000", "rgb(255,0,0)", "hsb(0,100,100)"
//
// Fade time can be specified by milliseconds, seconds, or minutes, e.g. "100ms", "1s", "1.5m", "now", "0s"
// Explicit numeric times take precedence over immediate words. Repeated numeric times
// must agree; conflicting, negative, or overflowing times return an error. A standalone
// zero is an immediate time, but a zero inside a color or LED selector is not.
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
		err := st.UnmarshalText([]byte(query))
		return st, err
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
			var rgb [3]uint8
			for i := range rgb {
				component, err := strconv.ParseUint(m[i+1], 10, 8)
				if err != nil {
					return nil, fmt.Errorf("invalid rgb color %q: %w", m[0], err)
				}
				rgb[i] = uint8(component)
			}
			return color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 0xff}, nil
		case "hsb":
			var hsb [3]float64
			for i := range hsb {
				component, err := strconv.ParseFloat(m[i+1], 64)
				if err != nil {
					return nil, fmt.Errorf("invalid hsb color %q: %w", m[0], err)
				}
				hsb[i] = component
			}
			return convRGBToColor(convHSBToRGB(hsb[0], hsb[1], hsb[2])), nil
		}
	}

	return nil, errNoColorMatch
}

func parseFadeTime(query string) (time.Duration, error) {
	var duration time.Duration
	found := false
	for _, pattern := range fadeTimeRegexPats {
		for _, match := range pattern.pat.FindAllStringSubmatchIndex(query, -1) {
			start, end := match[2], match[3]
			number := query[start:end]
			if invalidTimePrefix(query, start) || strings.HasPrefix(number, "-") {
				return 0, fmt.Errorf("b1: invalid fade time %q", query[match[0]:match[1]])
			}
			value, err := time.ParseDuration(number + pattern.unit)
			if err != nil {
				return 0, fmt.Errorf("b1: invalid fade time %q: %w", query[match[0]:match[1]], err)
			}
			if found && value != duration {
				return 0, errAmbiguousFade
			}
			duration, found = value, true
		}
	}
	// Do not interpret a color channel or LED number as a standalone zero time.
	for _, pat := range []*regexp.Regexp{colorRegexPats["rgb"], colorRegexPats["hsb"], ledIdxRegexPats[12]} {
		query = pat.ReplaceAllStringFunc(query, func(text string) string {
			return strings.Repeat(" ", len(text))
		})
	}
	foundZero := false
	for _, match := range zeroTimeRegexPat.FindAllStringIndex(query, -1) {
		if (match[0] > 0 && query[match[0]-1] == '.') || (match[1] < len(query) && query[match[1]] == '.') {
			continue
		}
		if invalidTimePrefix(query, match[0]) || query[match[0]] == '-' {
			return 0, fmt.Errorf("b1: invalid fade time %q", query[match[0]:match[1]])
		}
		foundZero = true
	}
	if found {
		// LightState text and the device use milliseconds. Preserve historical truncation.
		return duration.Truncate(time.Millisecond), nil
	}
	if foundZero || immediateRegexPat.MatchString(query) {
		return 0, nil
	}
	return 0, errNoFadeMatch
}

// A sign or decimal point before a matched number means the regex found a suffix
// of an invalid number (for example, "- 1s" or "1.2.3s"). Never use that suffix.
func invalidTimePrefix(query string, start int) bool {
	if start > 0 && query[start-1] == '.' {
		return true
	}
	prefix := strings.TrimRight(query[:start], " \t\r\n")
	if prefix == "" {
		return false
	}
	last := prefix[len(prefix)-1]
	return last == '-' || last == '+'
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
