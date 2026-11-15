package blink1_test

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	b1 "github.com/b1ug/blink1-go"
)

func BenchmarkParseTitle(b *testing.B) {
	q := `title: Crash Course in Go`
	b1.ParseTitle(q) // I've assumed you're calling ParseTitle function directly in the sample. Adjust as necessary.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b1.ParseTitle(q)
	}
}

func BenchmarkParseRepeatTimes(b *testing.B) {
	q := `will repeat 5 times infinitely`
	b1.ParseRepeatTimes(q)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b1.ParseRepeatTimes(q)
	}
}

func BenchmarkParseStateQuery_Simple(b *testing.B) {
	q := `led=2 color=pink fade=500ms`
	b1.ParseStateQuery(q)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b1.ParseStateQuery(q)
	}
}

func BenchmarkParseStateQuery_Complex(b *testing.B) {
	q := `fade to #add8e6 over 4.5s on led 0`
	b1.ParseStateQuery(q)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b1.ParseStateQuery(q)
	}
}

func TestParseTitle(t *testing.T) {
	tests := []struct {
		query    string
		expected string
		wantErr  bool
	}{
		{
			query:    "title: Crash Course in Go    ",
			expected: "Crash Course in Go",
		},
		{
			query:    "Title :  Crash Course in Rust  ",
			expected: "Crash Course in Rust",
		},
		{
			query:    "topic = Advanced Topics",
			expected: "Advanced Topics",
		},
		{
			query:    "topic= Another Topics",
			expected: "Another Topics",
		},
		{
			query:    "topic =Great Topics  ",
			expected: "Great Topics",
		},
		{
			query:    "TOPIC = LOVELY Topics  ",
			expected: "LOVELY Topics",
		},
		{
			query:    "idea: Revolutionize AI",
			expected: "Revolutionize AI",
		},
		{
			query:    "idea::: Revolutionize AI",
			expected: "Revolutionize AI",
		},
		{
			query:    "title=Deep Reinforcement Learning",
			expected: "Deep Reinforcement Learning",
		},
		{
			query:    "title No Borders",
			expected: "No Borders",
		},
		{
			query:    "subject: The Future of Quantum Computing",
			expected: "The Future of Quantum Computing",
		},
		{
			query:   "subj: The Future of Quantum Computing",
			wantErr: true,
		},
		{
			query:   "title = ",
			wantErr: true,
		},
		{
			query:   "topic:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := b1.ParseTitle(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTitle(%q) error = %v, wantErr = %v", tt.query, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got != tt.expected {
				t.Errorf("ParseTitle(%q) got = %q, want = %q", tt.query, got, tt.expected)
			}
		})
	}
}

func TestParseRepeatTimes(t *testing.T) {
	tests := []struct {
		query   string
		times   uint
		wantErr bool
	}{
		{
			query: "Repeat 25 times",
			times: 25,
		},
		{
			query: "repeat: 36",
			times: 36,
		},
		{
			query: "repeat=26",
			times: 26,
		},
		{
			query:   "repeat forever",
			times:   0,
			wantErr: false,
		},
		{
			query: "Repeat: 20 times",
			times: 20,
		},
		{
			query: "(repeat:10)",
			times: 10,
		},
		{
			query: "(pattern:name=LoveIsInTheAir, repeat=3)",
			times: 3,
		},
		{
			query: "repeat:0",
			times: 0,
		},
		{
			query: "Please repeat once",
			times: 1,
		},
		{
			query: "Repeat twice",
			times: 2,
		},
		{
			query: "Repeat thrice",
			times: 3,
		},
		{
			query: "repeat:always",
			times: 0,
		},
		{
			query: "repeat: infinitely",
			times: 0,
		},
		{
			query: "repeat: infinite",
			times: 0,
		},
		{
			query: "always repeat",
			times: 0,
		},
		{
			query: "infinite repeat",
			times: 0,
		},
		{
			query: "infinitely repeat",
			times: 0,
		},
		{
			query: "forever repeat",
			times: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := b1.ParseRepeatTimes(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepeatTimes(%q) error = %v, wantErr = %v", tt.query, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got != tt.times {
				t.Errorf("ParseRepeatTimes(%q) got = %v, want = %v", tt.query, got, tt.times)
			}
		})
	}
}

func TestParseTitleSingleCharacter(t *testing.T) {
	for _, title := range []string{"A", "灯", "💡"} {
		for _, prefix := range []string{"title: ", "topic=", "idea:::", "subject "} {
			query := prefix + title
			got, err := b1.ParseTitle(query)
			if err != nil || got != title {
				t.Fatalf("ParseTitle(%q): got %q err=%v, want %q", query, got, err, title)
			}
		}
	}
	for _, query := range []string{"title", "title:", "topic= ", "idea::: ", "subject  "} {
		if got, err := b1.ParseTitle(query); err == nil {
			t.Errorf("ParseTitle(%q): accepted empty title %q", query, got)
		}
	}
}

func TestLightStateUnmarshalTextBoundaries(t *testing.T) {
	const maxWholeMilliseconds = time.Duration(9223372036854) * time.Millisecond
	valid := []struct {
		text string
		led  b1.LEDIndex
		fade time.Duration
	}{
		{"#ff0000l0t0", b1.LEDAll, 0},
		{"#ff0000L1T1", b1.LED1, time.Millisecond},
		{"#ff0000L2T9223372036854", b1.LED2, maxWholeMilliseconds},
		{"#ff0000L01T001", b1.LED1, time.Millisecond},
	}
	for _, test := range valid {
		var got b1.LightState
		if err := got.UnmarshalText([]byte(test.text)); err != nil || got.LED != test.led || got.FadeTime != test.fade {
			t.Errorf("UnmarshalText(%q): got=%v err=%v, want led=%v fade=%v", test.text, got, err, test.led, test.fade)
		}
	}
	invalid := []string{
		"", "#ff0000L3T1", "#ff0000L9T1", "#ff0000L255T1", "#ff0000L-1T1",
		"#ff0000L9999999999999999999999T1", "#ff0000L1T-1", "#ff0000L1T-0",
		"#ff0000L1T9223372036855", "#ff0000L1T9223372036854775807",
		"#ff0000L1T9223372036854775808", "#ff0000L1T1garbage", "#ff0000L1T1;",
		"#ff0000M1T1", "#gg0000L1T1", "#ff0000 L1T1",
	}
	before := b1.LightState{Color: b1.ColorBlue, LED: b1.LED2, FadeTime: 123 * time.Millisecond}
	for _, text := range invalid {
		t.Run(text, func(t *testing.T) {
			got := before
			if err := got.UnmarshalText([]byte(text)); err == nil {
				t.Fatalf("UnmarshalText(%q): accepted invalid state %v", text, got)
			}
			if !reflect.DeepEqual(got, before) {
				t.Fatalf("UnmarshalText(%q): invalid input changed receiver from %v to %v", text, before, got)
			}
		})
	}
}

func TestParseStateQueryPreservesCompactValidationError(t *testing.T) {
	for _, test := range []struct {
		query   string
		wantErr string
	}{
		{"#ff0000L9T500", "invalid LED index"},
		{"#ff0000L1T-1", "invalid fade time"},
		{"#ff0000L1T9223372036854775807", "invalid fade time"},
	} {
		_, err := b1.ParseStateQuery(test.query)
		if err == nil || !strings.Contains(err.Error(), test.wantErr) {
			t.Errorf("ParseStateQuery(%q): err=%v, want containing %q", test.query, err, test.wantErr)
		}
	}
}

func TestStateSequenceInvalidEntryDoesNotChangeReceiver(t *testing.T) {
	before := b1.StateSequence{{Color: b1.ColorBlue, LED: b1.LED2, FadeTime: 123 * time.Millisecond}}
	got := append(b1.StateSequence(nil), before...)
	err := got.UnmarshalText([]byte("#ff0000L1T5;#ff0000L9T5"))
	if err == nil || !reflect.DeepEqual(got, before) {
		t.Fatalf("invalid sequence changed receiver: got=%v err=%v, want %v", got, err, before)
	}
}

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

// The public SDK parser returns LightState directly, not an intermediate DTO.
func ExampleParseStateQuery_commands() {
	for _, command := range []string{"set red", "off on led 2", "fade to blue over 2s on led 1"} {
		state, err := b1.ParseStateQuery(command)
		if err != nil {
			panic(err)
		}
		fmt.Println(state)
	}

	// Output:
	// 🎨(color=#FF0000 led=0 fade=0s)
	// 🎨(color=#000000 led=2 fade=0s)
	// 🎨(color=#0000FF led=1 fade=2s)
}

func ExampleParseStateQuery_fields() {
	state, err := b1.ParseStateQuery("led=1 color=hsv(210,75,90) fade=20ms")
	if err != nil {
		panic(err)
	}
	text, err := state.MarshalText()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(text))

	// Output:
	// #398FE6L1T20
}

// Validate the complete command before opening the hardware. Once a controller
// is open, the parsed state can be passed to PlayState without conversion.
// This hardware example is compiled, but has no Output comment and is not run.
func ExampleController_PlayState_query() {
	state, err := b1.ParseStateQuery("fade to blue over 2s on led 1")
	if err != nil {
		panic(err)
	}
	controller, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer controller.Close()
	if err := controller.PlayState(state); err != nil {
		panic(err)
	}
}
