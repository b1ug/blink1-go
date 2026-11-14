package blink1_test

import (
	"errors"
	"image/color"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	b1 "github.com/b1ug/blink1-go"
)

func TestParseColorCompleteExpressions(t *testing.T) {
	for _, tc := range []struct {
		input string
		hex   string
	}{
		{"red", "#FF0000"}, {" GREEN ", "#008000"}, {"lime", "#00FF00"},
		{"plum", "#DDA0DD"}, {"violet", "#EE82EE"}, {"rebeccapurple", "#663399"},
		{"#aBc", "#AABBCC"}, {"#12abEf", "#12ABEF"},
		{"rgb( 1 , 2 , 3 )", "#010203"}, {"RGB (255,0,255)", "#FF00FF"},
		{"hsb(120,100,100)", "#00FF00"}, {"HSV(210,75,90)", "#398FE6"},
		{"hsv(360,100,100)", "#FF0000"}, {"hsb(0,0,50)", "#808080"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := b1.ParseColor(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := got.(color.RGBA); !ok {
				t.Fatalf("got %T, want the existing SDK color.RGBA representation", got)
			}
			if hex := b1.ColorToHex(got); hex != tc.hex {
				t.Fatalf("got %s, want %s", hex, tc.hex)
			}
			// The parsed value already satisfies every SDK color-taking API.
			state := b1.NewLightState(got, 20*time.Millisecond, b1.LED1)
			if state.Color != got || state.LED != b1.LED1 {
				t.Fatal("color required a caller-side conversion")
			}
		})
	}
}

func TestParseColorRejectsPartialAndClippedInput(t *testing.T) {
	for _, input := range []string{
		"", "please red", "red blue", "off", "on", "@mint", "mint", "transparent",
		"abc", "#abcd", "#aabbccdd", "#abc trailing", "rgb(1,2,3) extra",
		"rgb(256,0,0)", "rgb(-1,2,3)", "rgb(1.5,2,3)", "rgb(1,2,3,4)",
		"hsb(120,101,100)", "hsv(361,100,100)", "hsv(-1,100,100)",
		"hsv(120,100,-1)", "hsv(120,100,100.000001)", "hsv(1e2,100,100)",
		"hsv(1.0000001,100,100)", "hsv(NaN,100,100)", "hsv(Inf,100,100)",
	} {
		t.Run(input, func(t *testing.T) {
			got, err := b1.ParseColor(input)
			if err == nil || got != nil {
				t.Fatalf("accepted partial color: %#v, %v", got, err)
			}
			var diagnostic *b1.ParseError
			if !errors.As(err, &diagnostic) || diagnostic.Code == "" {
				t.Fatalf("missing inspectable diagnostic: %v", err)
			}
		})
	}
}

func TestParseStateCanonicalCommands(t *testing.T) {
	for _, tc := range []struct {
		input string
		hex   string
		led   b1.LEDIndex
		fade  time.Duration
	}{
		{"set red", "#FF0000", b1.LEDAll, 0},
		{"SET #aBc now", "#AABBCC", b1.LEDAll, 0},
		{"set led 1 to red", "#FF0000", b1.LED1, 0},
		{"set blue on led 2", "#0000FF", b1.LED2, 0},
		{"set green on led all", "#008000", b1.LEDAll, 0},
		{"off", "#000000", b1.LEDAll, 0},
		{"off on led 2 now", "#000000", b1.LED2, 0},
		{"fade to blue over 2s on led 1", "#0000FF", b1.LED1, 2 * time.Second},
		{"fade led 2 hsv(210,75,90) in 0.02s", "#398FE6", b1.LED2, 20 * time.Millisecond},
		{"fade red over 655350ms", "#FF0000", b1.LEDAll, 655350 * time.Millisecond},
		{"led=1 color=red fade=1s", "#FF0000", b1.LED1, time.Second},
		{"fade=0ms color=rgb(1, 2, 3) led=2", "#010203", b1.LED2, 0},
		{"color=#abc led=0 fade=10ms", "#AABBCC", b1.LEDAll, 10 * time.Millisecond},
		{"set red // leave the other light alone", "#FF0000", b1.LEDAll, 0},
		{"#FF0000L1T20", "#FF0000", b1.LED1, 20 * time.Millisecond},
		{" #aabbccL2T0 // compact data", "#AABBCC", b1.LED2, 0},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := b1.ParseStateQuery(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if hex := b1.ColorToHex(got.Color); hex != tc.hex || got.LED != tc.led || got.FadeTime != tc.fade {
				t.Fatalf("got %s, want color=%s led=%d fade=%v", got, tc.hex, tc.led, tc.fade)
			}
			if _, ok := got.Color.(color.RGBA); !ok {
				t.Fatalf("got %T, want SDK color.RGBA", got.Color)
			}
		})
	}
}

func TestParseStateRejectsUnknownIntentAndReturnsNoPartialState(t *testing.T) {
	for _, input := range []string{
		"", "led 1 red 1s", "do not turn led 1 red now", "not set red", "set not red",
		"set red trailing", "set red 再蓝", "set red blue", "set red on led 1 on led 2",
		"set led 1 red on led 1", "set red hold 1s", "set red for 3s",
		"set red then off", "blink led 1 red 1s", "orbit red blue", "rainbow",
		"off over 2s", "off red", "fade blue", "fade blue over 0ms",
		"fade blue over 1s now", "fade blue over 1s over 1000ms",
		"fade blue over 15ms", "fade blue over 0.001s", "fade blue over -1s",
		"fade blue over +1s", "fade blue over 655360ms", "fade blue over 9223372036854775808ms",
		"set red on led 3", "set red on led top", "set red on led bottom",
		"led=1 color=red", "led=1 color=red fade=1s extra=2",
		"led=1 color=red fade=1s led=1", "set color=red led=1 fade=0ms",
		"led=1 color=red fade=15ms", "led=1 color=red fade=-1ms",
		"#ff0000L1T15", "#ff0000L1T655360", "#ff0000L9T20",
		"set red\noff", "set red\r\noff", string([]byte{'s', 'e', 't', ' ', 0xff}),
		strings.Repeat("x", 4097),
	} {
		t.Run(input, func(t *testing.T) {
			state, err := b1.ParseStateQuery(input)
			if err == nil || state.Color != nil || state.LED != 0 || state.FadeTime != 0 {
				t.Fatalf("returned executable partial state for %q: %#v, %v", input, state, err)
			}
			var diagnostic *b1.ParseError
			if !errors.As(err, &diagnostic) || diagnostic.Code == "" || diagnostic.Hint == "" {
				t.Fatalf("missing actionable typed error: %v", err)
			}
			if diagnostic.Start < 0 || diagnostic.End < diagnostic.Start || diagnostic.End > len(input) {
				t.Fatalf("invalid original source range: %#v", diagnostic)
			}
			if utf8.ValidString(input) && !strings.ContainsAny(input, "\r\n") && diagnostic.Column != utf8.RuneCountInString(input[:diagnostic.Start])+1 {
				t.Fatalf("byte/rune column mismatch: %#v", diagnostic)
			}
		})
	}
}

func TestParseStateDeterministicConflicts(t *testing.T) {
	const input = "set red on led 1 on led 2"
	_, first := b1.ParseStateQuery(input)
	if first == nil {
		t.Fatal("accepted conflicting targets")
	}
	for n := 0; n < 256; n++ {
		state, err := b1.ParseStateQuery(input)
		if err == nil || err.Error() != first.Error() || state.Color != nil {
			t.Fatalf("non-deterministic parse at %d: %#v %v", n, state, err)
		}
	}
}

func TestCompactCodecIsDataAndParserChecksExecutability(t *testing.T) {
	var data b1.LightState
	if err := data.UnmarshalText([]byte("#ff0000L1T15")); err != nil || data.FadeTime != 15*time.Millisecond {
		t.Fatalf("data codec unexpectedly changed: %#v %v", data, err)
	}
	if state, err := b1.ParseStateQuery("#ff0000L1T15"); err == nil || state.Color != nil {
		t.Fatalf("unrepresentable executable state: %#v %v", state, err)
	}
}

func TestRejectedStateIsSafeToDescribeButCannotBeEncoded(t *testing.T) {
	state, err := b1.ParseStateQuery("do not set red")
	if err == nil || state.Color != nil {
		t.Fatalf("expected no executable state: %#v %v", state, err)
	}
	if description := state.String(); description != "🎨(color=<nil> led=0 fade=0s)" {
		t.Fatalf("unsafe or misleading zero-state description: %s", description)
	}
	if text, err := state.MarshalText(); err == nil || text != nil {
		t.Fatalf("encoded missing color as a valid state: %q %v", text, err)
	}
}
