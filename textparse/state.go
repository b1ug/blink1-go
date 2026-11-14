package textparse

import (
	"strings"
	"time"

	"github.com/b1ug/colorkit"
)

const maxFade = 655350 * time.Millisecond

// State is one target color, an LED selector (0=all, 1 or 2), and the time
// taken to reach the target. It cannot express holds, repetition, or sequences.
type State struct {
	Color    colorkit.RGB8
	LED      uint8
	FadeTime time.Duration
}

// ParseState parses one complete set/off/fade command or the explicit
// led=... color=... fade=... form. Keywords are ASCII case-insensitive.
// Commands default to LED0; field form requires all three keys exactly once.
// A // trailing comment may begin at the start of the line or after whitespace.
// The entire original input, including comments, must satisfy the UTF-8,
// single-line, and 4096-byte limits before the comment is ignored.
// Every failure returns a zero State.
func ParseState(input string) (State, error) {
	c, err := scan(input, true)
	if err != nil {
		return State{}, err
	}
	if len(c.tokens) > 1 && c.tokens[1].text == "=" {
		return c.fields()
	}
	return c.command()
}

func (c *cursor) command() (State, error) {
	action := c.next()
	if action.text == "" {
		return State{}, c.err(action, "missing_field", "action", "an action is required", "Use set, off, fade, or the explicit three-key state form.", nil)
	}
	kind := lowerASCII(action.text)
	if kind != "set" && kind != "off" && kind != "fade" {
		return State{}, c.err(action, "unsupported_intent", "action", "unsupported single-state action", "Use set/off/fade here; use nb1 for higher-level actions or sequences.", nil)
	}
	var state State
	hasLED := false
	if kind == "off" {
		if c.take("on") {
			if err := c.expect("led", "led"); err != nil {
				return State{}, err
			}
			hasLED = true
			led, err := c.led()
			if err != nil {
				return State{}, err
			}
			state.LED = led
		} else if c.take("led") {
			hasLED = true
			led, err := c.led()
			if err != nil {
				return State{}, err
			}
			state.LED = led
		}
	} else {
		if c.take("led") {
			hasLED = true
			led, err := c.led()
			if err != nil {
				return State{}, err
			}
			state.LED = led
		}
		c.take("to")
		if lowerASCII(c.peek().text) == "led" {
			return State{}, c.err(c.peek(), "conflicting_values", "led", "repeated LED selector", "Specify the LED only once.", nil)
		}
		color, err := c.color()
		if err != nil {
			return State{}, err
		}
		state.Color = color
		if kind == "fade" {
			if !c.take("over") && !c.take("in") {
				code := "unexpected_input"
				if c.end() {
					code = "missing_field"
				}
				return State{}, c.err(c.peek(), code, "fade", "fade requires an explicit transition duration", "Write fade COLOR over 2s or fade COLOR in 2s.", nil)
			}
			d, err := c.duration(false)
			if err != nil {
				return State{}, err
			}
			state.FadeTime = d
		}
	}
	if c.take("on") {
		selector := c.tokens[c.pos-1]
		if hasLED {
			return State{}, c.err(selector, "conflicting_values", "led", "LED selector was specified more than once", "Use either a leading or a trailing selector.", nil)
		}
		if err := c.expect("led", "led"); err != nil {
			return State{}, err
		}
		led, err := c.led()
		if err != nil {
			return State{}, err
		}
		state.LED = led
	}
	if lowerASCII(c.peek().text) == "now" {
		if kind == "fade" {
			return State{}, c.err(c.peek(), "conflicting_values", "fade", "now conflicts with a positive fade duration", "Use set for an immediate change, or omit now.", nil)
		}
		c.next()
	}
	if !c.end() {
		return State{}, c.trailing()
	}
	return state, nil
}

func (c *cursor) fields() (State, error) {
	var state State
	seen := make(map[string]bool, 3)
	for !c.end() {
		keyToken := c.next()
		key := lowerASCII(keyToken.text)
		if key != "led" && key != "color" && key != "fade" {
			return State{}, c.err(keyToken, "unexpected_input", "input", "only led, color, and fade fields are accepted", "Supply exactly led=... color=... fade=....", nil)
		}
		if seen[key] {
			return State{}, c.err(keyToken, "conflicting_values", key, "field was specified more than once", "Supply each of led/color/fade once.", nil)
		}
		if err := c.expect("=", key); err != nil {
			return State{}, err
		}
		if c.end() || c.pos+1 < len(c.tokens) && c.tokens[c.pos+1].text == "=" {
			return State{}, c.err(c.peek(), "missing_field", key, "field requires a value", "Supply each of led/color/fade once.", nil)
		}
		seen[key] = true
		switch key {
		case "led":
			led, err := c.led()
			if err != nil {
				return State{}, err
			}
			state.LED = led
		case "color":
			color, err := c.color()
			if err != nil {
				return State{}, err
			}
			state.Color = color
		case "fade":
			d, err := c.duration(true)
			if err != nil {
				return State{}, err
			}
			state.FadeTime = d
		}
	}
	for _, key := range []string{"led", "color", "fade"} {
		if !seen[key] {
			return State{}, c.err(c.peek(), "missing_field", key, "explicit state form requires all three fields", "Supply led=... color=... fade=....", nil)
		}
	}
	return state, nil
}

func (c *cursor) led() (uint8, error) {
	t := c.next()
	switch lowerASCII(t.text) {
	case "0", "all", "both":
		return 0, nil
	case "1", "first":
		return 1, nil
	case "2", "second":
		return 2, nil
	case "":
		return 0, c.err(t, "missing_field", "led", "an LED selector is required", "Use 0, 1, 2, all, both, first, or second.", nil)
	default:
		return 0, c.err(t, "value_out_of_range", "led", "unsupported LED selector", "Use 0, 1, 2, all, both, first, or second; physical positions are not inferred.", nil)
	}
}

func (c *cursor) duration(allowZero bool) (time.Duration, error) {
	t := c.next()
	if t.text == "" {
		return 0, c.err(t, "missing_field", "fade", "a transition duration is required", "Use a value such as 10ms, 2s, or 1.5m.", nil)
	}
	text := lowerASCII(t.text)
	unit := ""
	for _, suffix := range []string{"ms", "s", "m"} {
		if strings.HasSuffix(text, suffix) {
			unit = suffix
			break
		}
	}
	if unit == "" || !durationNumber(text[:len(text)-len(unit)]) {
		return 0, c.err(t, "invalid_duration", "fade", "duration requires an unsigned decimal and ms, s, or m", "Use at most six fractional digits without signs or exponent notation.", nil)
	}
	d, err := time.ParseDuration(text)
	if err != nil {
		return 0, c.err(t, "value_out_of_range", "fade", "duration cannot be represented", "Use an exact 10ms multiple not exceeding 655350ms.", err)
	}
	if !allowZero && d == 0 {
		return 0, c.err(t, "value_out_of_range", "fade", "fade requires a positive duration", "Use set/off for an immediate change.", nil)
	}
	if err := ValidateFade(d); err != nil {
		var detail = err.(*Error)
		return 0, c.err(t, detail.Code, detail.Field, detail.Message, detail.Hint, detail.Cause)
	}
	return d, nil
}

func durationNumber(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		return digits(s)
	}
	return len(parts) == 2 && digits(parts[0]) && digits(parts[1]) && len(parts[1]) <= 6
}

// ValidateFade checks the SDK's single-state protocol range and precision.
// Zero is an immediate state. Positive values must be exact 10ms multiples
// no greater than 655350ms. The error has no source span; parsing attaches it.
func ValidateFade(d time.Duration) error {
	if d < 0 || d > maxFade {
		return &Error{Code: "value_out_of_range", Field: "fade", Line: 1, Column: 1,
			Message: "fade duration must be within 0..655350ms", Hint: "Use set/off for zero, or a positive duration within the protocol limit."}
	}
	if d%(10*time.Millisecond) != 0 {
		return &Error{Code: "invalid_duration", Field: "fade", Line: 1, Column: 1,
			Message: "fade duration must use exact 10ms units", Hint: "Choose an exact 10ms multiple; the value will not be rounded."}
	}
	return nil
}
