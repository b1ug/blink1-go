package blink1

import (
	"fmt"
	"image/color"
	"strconv"
	"time"
)

// LEDIndex represents LED on the blink(1) device
type LEDIndex byte

const (
	// LEDAll represents all LEDs on the blink(1) device
	LEDAll LEDIndex = iota
	// LED1 represents the first LED on the blink(1) device, usually the top one, with 'blink(1)' label
	LED1
	// LED2 represents the second LED on the blink(1) device, usually the bottom one, with 'ThingM' logo
	LED2
)

// ToByte converts LEDType to byte, and returns 0 if the LEDType is invalid.
func (l LEDIndex) ToByte() byte {
	if l < LEDAll || l > LED2 {
		return byte(LEDAll)
	}
	return byte(l)
}

// String returns a string representation of LEDIndex.
func (l LEDIndex) String() string {
	switch l {
	case LED1:
		return "LED 1"
	case LED2:
		return "LED 2"
	default:
		return "All LED"
	}
}

// DevicePatternState is a blink(1) pattern playing state for low-level APIs.
type DevicePatternState struct {
	IsPlaying    bool // Is playing
	CurrentPos   uint // Current position
	LoopStartPos uint // Loop start position, inclusive
	LoopEndPos   uint // Loop end position, exclusive
	RepeatTimes  uint // Remaining times to repeat
}

func (st DevicePatternState) String() string {
	return fmt.Sprintf("%s{playing=%t cur=%d loop=[%d,%d) left=%d}", convPlayingToEmoji(st.IsPlaying), st.IsPlaying, st.CurrentPos, st.LoopStartPos, st.LoopEndPos, st.RepeatTimes)
}

// PatternState represents a blink(1) pattern playing state for high-level APIs.
type PatternState struct {
	IsPlaying       bool // Is playing
	CurrentPosition uint // Current position
	StartPosition   uint // Loop start position, inclusive
	EndPosition     uint // Loop end position, exclusive
	RepeatTimes     uint // Remaining times to repeat
}

func (st PatternState) String() string {
	return fmt.Sprintf("%s(playing=%t cur=%d loop=[%d,%d) left=%d)", convPlayingToEmoji(st.IsPlaying), st.IsPlaying, st.CurrentPosition, st.StartPosition, st.EndPosition, st.RepeatTimes)
}

// DeviceLightState is a blink(1) light state for low-level APIs.
type DeviceLightState struct {
	R, G, B      byte     // RGB values
	LED          LEDIndex // Which LED to address (0=all, 1=1st LED, 2=2nd LED)
	FadeTimeMsec uint     // Fade time in milliseconds
}

func (st DeviceLightState) String() string {
	return fmt.Sprintf("🎨{color=#%02X%02X%02X led=%d fade=%dms}", st.R, st.G, st.B, st.LED, st.FadeTimeMsec)
}

// LightState is a blink(1) light state for high-level APIs.
type LightState struct {
	Color    color.Color   // Color to set
	LED      LEDIndex      // Which LED to address (0=all, 1=1st LED, 2=2nd LED)
	FadeTime time.Duration // Fade time to state
}

func (st LightState) String() string {
	return fmt.Sprintf("🎨(color=%s led=%d fade=%v)", convColorToHex(st.Color), st.LED, st.FadeTime)
}

// Pattern is a sequence of LightState to play on blink(1).
type Pattern struct {
	StartPosition uint         // Loop start position, inclusive
	EndPosition   uint         // Loop end position, inclusive
	RepeatTimes   uint         // How many times to repeat, 0 means infinite
	States        []LightState // Slice of states to execute in pattern, non-empty patterns will be set to the device automatically
}

func (p Pattern) String() string {
	var repeat string
	if p.RepeatTimes == 0 {
		repeat = "∞"
	} else {
		repeat = strconv.Itoa(int(p.RepeatTimes))
	}
	return fmt.Sprintf("🎼(loop=[%d,%d] repeat=%s states=%d)", p.StartPosition, p.EndPosition, repeat, len(p.States))
}
