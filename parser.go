package blink1

import (
	"errors"
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/b1ug/blink1-go/textparse"
)

var (
	regexOnce         sync.Once
	titleRegexPat     *regexp.Regexp
	repeatRegexPat    *regexp.Regexp
	stateTextRegexPat *regexp.Regexp
	emptyStr          string
	errNoTitleMatch   = errors.New("b1: no title match")
	errNoRepeatMatch  = errors.New("b1: no repeat times match")
)

func initRegex() {
	repeatRegexPat = regexp.MustCompile(`\brepeat\s*[:=]*\s*(\d+|\bonce|\btwice|\bthrice|\bforever|\balways|\binfinite(?:ly)?)\b|\b(infinite(?:ly)?|forever|always|once|twice|thrice)\s+repeat\b`)
	titleRegexPat = regexp.MustCompile(`(?i)\b(title|topic|idea|subject)\s*[:=]*\s*`)
	stateTextRegexPat = regexp.MustCompile(`(?i)^#([0-9a-f]{6})L([+-]?\d+)T([+-]?\d+)$`)
}

// ParseError describes a failed complete color or state expression. Code and
// Field are stable machine-readable values; Start and End address a half-open
// byte range in the original input. Use errors.As to inspect the diagnostic.
// A parser error never accompanies an executable partial LightState.
type ParseError = textparse.Error

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

// ParseColor parses one complete color expression: an opaque CSS name,
// #rgb/#rrggbb, rgb(r,g,b), or hsv/hsb(h,s,v). HSV hue uses degrees [0,360]
// and saturation/value use percentages [0,100], with at most six decimals.
// It rejects unknown surrounding content, clipping, on/off aliases and alpha.
// The returned color can be passed directly to NewLightState or PlayColor.
func ParseColor(query string) (color.Color, error) {
	rgb, err := textparse.ParseColor(query)
	if err != nil {
		return nil, err
	}
	return convRGBToColor(rgb.R, rgb.G, rgb.B), nil
}

// ParseStateQuery parses one complete Set, Off, or Fade state. For example:
//
//	set red
//	off on led 2
//	fade to blue over 2s on led 1
//	led=1 color=red fade=1s
//
// Set and Off default to LEDAll and immediate application. Fade requires a
// positive transition in exact 10ms units, up to 655350ms. The field form
// requires exactly led, color and fade, once each; fade=0ms is immediate.
// Explicit compact state data (#RRGGBBLnTms) is decoded and checked against
// the same executable fade limits. Its standalone data codec remains separate.
// Higher-level effects, holding, negation, multiple steps and unknown trailing
// content are rejected. Errors return a zero LightState with no color.
// Parsing does not open a device or access local configuration.
func ParseStateQuery(query string) (LightState, error) {
	// Compact state text is a supported explicit data format, not a fallback to
	// the previous field extractor. The pure parser diagnoses malformed input.
	if len(query) <= 4096 && utf8.ValidString(query) && !strings.ContainsAny(query, "\r\n\v\f\u0085\u2028\u2029") {
		body := query
		if comment := strings.Index(body, "//"); comment >= 0 {
			previous, _ := utf8.DecodeLastRuneInString(body[:comment])
			if comment == 0 || unicode.IsSpace(previous) {
				body = body[:comment]
			}
		}
		body = strings.TrimSpace(body)
		regexOnce.Do(initRegex)
		if stateTextRegexPat.MatchString(body) {
			var state LightState
			if err := state.UnmarshalText([]byte(body)); err != nil {
				return LightState{}, compactParseError(query, body, "invalid_input", "state", err, "Use #RRGGBBL{0,1,2}T{milliseconds}.")
			}
			if err := textparse.ValidateFade(state.FadeTime); err != nil {
				return LightState{}, compactParseError(query, body, "invalid_duration", "fade", err, "Use 0ms for immediate application or a representable fade duration.")
			}
			return state, nil
		}
	}
	s, err := textparse.ParseState(query)
	if err != nil {
		return LightState{}, err
	}
	return NewLightState(convRGBToColor(s.Color.R, s.Color.G, s.Color.B), s.FadeTime, LEDIndex(s.LED)), nil
}

func compactParseError(input, body, code, field string, cause error, hint string) error {
	start := strings.Index(input, body)
	if start < 0 {
		start = 0
	}
	return &ParseError{Code: code, Field: field, Start: start, End: start + len(body), Line: 1, Column: utf8.RuneCountInString(input[:start]) + 1, Message: cause.Error(), Hint: hint, Cause: cause}
}
