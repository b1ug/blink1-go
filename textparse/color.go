package textparse

import (
	"errors"
	"strconv"

	"github.com/b1ug/colorkit"
)

// ParseColor accepts exactly one CSS named color, #rgb/#rrggbb literal, or
// rgb/hsb/hsv function. Function components may have surrounding whitespace.
// RGB requires integers in [0,255]; HSV/HSB uses degrees and percentages with
// at most six decimal places. No input is clipped or interpreted as an action.
// Unlike ParseState, ParseColor does not accept trailing comments.
func ParseColor(input string) (colorkit.RGB8, error) {
	c, err := scan(input, false)
	if err != nil {
		return colorkit.RGB8{}, err
	}
	color, err := c.color()
	if err != nil {
		return colorkit.RGB8{}, err
	}
	if !c.end() {
		return colorkit.RGB8{}, c.trailing()
	}
	return color, nil
}

func (c *cursor) color() (colorkit.RGB8, error) {
	t := c.next()
	if t.text == "" {
		return colorkit.RGB8{}, c.err(t, "missing_field", "color", "a color is required", "Provide a CSS color, Hex literal, or rgb/hsb/hsv function.", nil)
	}
	kind := lowerASCII(t.text)
	if kind != "rgb" && kind != "hsb" && kind != "hsv" {
		value, err := colorkit.Parse(t.text)
		if err != nil {
			return colorkit.RGB8{}, c.err(t, "invalid_color", "color", "unsupported or malformed color", "Use one complete CSS name, Hex literal, or rgb/hsb/hsv function.", err)
		}
		return value, nil
	}
	if err := c.expect("(", "color"); err != nil {
		return colorkit.RGB8{}, err
	}
	var parts [3]token
	for i := range parts {
		parts[i] = c.next()
		if parts[i].text == "" || parts[i].text == "," || parts[i].text == ")" || parts[i].text == "(" || parts[i].text == "=" {
			return colorkit.RGB8{}, c.err(parts[i], "missing_field", "color", "three numeric color components are required", "Separate exactly three components with commas.", nil)
		}
		separator := ","
		if i == 2 {
			separator = ")"
		}
		if err := c.expect(separator, "color"); err != nil {
			return colorkit.RGB8{}, err
		}
	}
	if kind == "rgb" {
		var channels [3]uint8
		fields := [3]string{"color.r", "color.g", "color.b"}
		for i, part := range parts {
			if !digits(part.text) {
				return colorkit.RGB8{}, c.err(part, "invalid_color", fields[i], "RGB components must be nonnegative integers", "Use integer channels from 0 to 255.", colorkit.ErrSyntax)
			}
			v, err := strconv.ParseUint(part.text, 10, 8)
			if err != nil {
				return colorkit.RGB8{}, c.err(part, "value_out_of_range", fields[i], "RGB component is outside 0 to 255", "Use integer channels from 0 to 255.", colorkit.ErrRange)
			}
			channels[i] = uint8(v)
		}
		return colorkit.RGB8{R: channels[0], G: channels[1], B: channels[2]}, nil
	}
	value, err := colorkit.FromHSVDecimal(parts[0].text, parts[1].text, parts[2].text)
	if err != nil {
		idx, field := 0, "color.h"
		var detail *colorkit.Error
		if errors.As(err, &detail) {
			switch detail.Field {
			case "s":
				idx, field = 1, "color.s"
			case "v":
				idx, field = 2, "color.v"
			}
		}
		code := "invalid_color"
		if errors.Is(err, colorkit.ErrRange) {
			code = "value_out_of_range"
		}
		return colorkit.RGB8{}, c.err(parts[idx], code, field, "invalid HSV/HSB component", "Use H in 0..360 and S/V in 0..100, with at most six decimal places.", err)
	}
	return value, nil
}

func digits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func (c *cursor) trailing() error {
	t := c.peek()
	word := lowerASCII(t.text)
	switch word {
	case "not", "never", "don't", "cannot", "no":
		return c.err(t, "unsupported_intent", "action", "negation is not supported", "State the desired supported command explicitly.", nil)
	case "hold", "for", "then", "blink", "breathe", "pulse", "alternate", "orbit", "rainbow":
		return c.err(t, "unsupported_intent", "action", "this intent cannot be represented by one state", "Use the nb1 Action or Sequence interface.", nil)
	case "on", "led":
		return c.err(t, "conflicting_values", "led", "unexpected or repeated LED selector", "Specify one LED selector, before or after the color.", nil)
	case "now", "over", "in":
		return c.err(t, "conflicting_values", "fade", "unexpected or repeated time declaration", "Use set/off for immediate output, or fade with one transition duration.", nil)
	}
	if _, ok := colorkit.LookupName(t.text); ok || word == "rgb" || word == "hsb" || word == "hsv" || len(word) > 0 && word[0] == '#' {
		return c.err(t, "conflicting_values", "color", "more than one color was supplied", "Use one color per state; use nb1 for a sequence or paired colors.", nil)
	}
	return c.err(t, "unexpected_input", "input", "unexpected input", "Remove unsupported words or use a documented Action/Sequence entry.", nil)
}
