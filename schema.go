package blink1

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"
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

// LightState is a composite data type that represents the state of a light in a blink(1) high-level APIs context. It includes three fields: Color, LED, and FadeTime.
//
// The Color field, a value of type color.Color, represents the color to be set for the LightState.
// The LED field, a value of type LEDIndex, represents which LED to address. Here, 0 represents all LEDs, 1 represents the 1st LED, and 2 represents the 2nd LED.
// The FadeTime field, a value of type time.Duration, represents the fade time to the LightState, specified in duration.
//
// This struct is serialized in string format with three components, separated by specific letters for ease of understanding and clean formatting:
//    #RRGGBBL{0,1,2}T{fade time in milliseconds}
//
//    1. Color is converted to hexadecimal (HEX) representation using convColorToHex function;
//    2. LED is represented by its corresponding LEDIndex value prefixed by 'L';
//    3. FadeTime is represented by its millisecond value prefixed by 'T';
//
// For example, a reddish color (Hex #FF0000), targeting LED 1, with a fade time of 200ms would be serialized as "#FF0000L1T200".
type LightState struct {
	Color    color.Color   // Color to set
	LED      LEDIndex      // Which LED to address (0=all, 1=1st LED, 2=2nd LED)
	FadeTime time.Duration // Fade time to state
}

// MarshalText implements the encoding.TextMarshaler interface.
func (st LightState) MarshalText() (text []byte, err error) {
	s := fmt.Sprintf(`%sL%dT%d`,
		convColorToHex(st.Color),
		st.LED,
		st.FadeTime.Milliseconds())
	return []byte(s), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (st *LightState) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return errEmptyState
	}
	// parse
	var (
		s             = string(text)
		r, g, b       uint8
		led           LEDIndex
		fadeTimeMilli int
	)
	if _, err := fmt.Sscanf(strings.ToUpper(s), "#%02X%02X%02XL%dT%d", &r, &g, &b, &led, &fadeTimeMilli); err != nil {
		return fmt.Errorf("invalid format for LightState: %w", err)
	}
	// fill in
	*st = LightState{}
	st.Color = convRGBToColor(r, g, b)
	st.LED = led
	st.FadeTime = time.Duration(fadeTimeMilli) * time.Millisecond
	return nil
}

// String method is used to represent a LightState in string format for human comprehension,
// annotating each element by its field name, and using emojis and symbols for clarity and aesthetics.
// For example: 🎨(color=#FF0000 led=1 fade=200ms)
func (st LightState) String() string {
	return fmt.Sprintf("🎨(color=%s led=%d fade=%v)", convColorToHex(st.Color), st.LED, st.FadeTime)
}

var (
	stateSeqSeparator = ";"
	errEmptyState     = errors.New("empty state can't be deserialized")
)

// StateSequence is a data type that represents a sequence of LightState, used to define the sequence of states to be played on the blink(1) device. The type is essentially a slice of LightState.
type StateSequence []LightState

// String method is used to output the StateSequence in a human-readable string format.
//
// If the sequence is empty, it returns "🔄(empty)". This means that there are no light states in this sequence.
//
// If the length of the sequence is exactly 1, it simply outputs the lone LightState in the sequence, as there are no additional states in the sequence to specify. For example, "🔄(#FF0000L1T200)" would indicate a single state of reddish color LED1 with a fade time of 200ms as the sole state in the sequence.
//
// If there is more than one LightState in the sequence, it prints out the first LightState in the sequence and the count of LightStates, separated by dots. For example, "🔄(#FF0000L1T200...5)" indicates a sequence of 5 states, starting with a reddish color LED1 with a fade time of 200ms. This is the default case when there are more than one states in the sequence.
func (seq StateSequence) String() string {
	c := len(seq)
	if c == 0 {
		return "🔄(empty)"
	} else if c == 1 {
		return fmt.Sprintf("🔄(%v)", seq[0])
	}
	return fmt.Sprintf("🔄(%v...%d)", seq[0], c)
}

// MarshalText implements the encoding.TextMarshaler interface.
// This method will convert a StateSequence into a string of semicolon-separated state sequences.
// However, the individual states are expected to be marshalled using the LightState MarshalText method before joining
// together into a single string.
// For example, a sequence of two states might be serialized as "#FF0000L1T200;#00FF00L2T300".
func (seq StateSequence) MarshalText() (text []byte, err error) {
	ls := make([]string, len(seq))
	for i, st := range seq {
		b, _ := st.MarshalText()
		ls[i] = string(b)
	}
	return []byte(strings.Join(ls, stateSeqSeparator)), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (seq *StateSequence) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*seq = StateSequence{}
		return nil
	}
	// split
	ls := strings.Split(string(text), stateSeqSeparator)
	// parse and fill in
	ss := make(StateSequence, len(ls))
	for i, l := range ls {
		if err := ss[i].UnmarshalText([]byte(l)); err != nil {
			return err
		}
	}
	// write back
	*seq = ss
	return nil
}

// Pattern is a sequence of LightState to play on blink(1).
type Pattern struct {
	StartPosition uint          // Loop start position, inclusive
	EndPosition   uint          // Loop end position, inclusive
	RepeatTimes   uint          // How many times to repeat, 0 means infinite
	Sequence      StateSequence // Sequence of states to execute in pattern, non-empty patterns will be set to the device automatically
}

func (p Pattern) String() string {
	var repeat string
	if p.RepeatTimes == 0 {
		repeat = "∞"
	} else {
		repeat = strconv.Itoa(int(p.RepeatTimes))
	}
	return fmt.Sprintf("🎼(loop=[%d,%d] repeat=%s seq=%d)", p.StartPosition, p.EndPosition, repeat, len(p.Sequence))
}
