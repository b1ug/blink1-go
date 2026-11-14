package textparse_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/b1ug/blink1-go/textparse"
	"github.com/b1ug/colorkit"
)

func TestColorLiterals(t *testing.T) {
	cases := []struct{ input, hex string }{
		{"red", "#ff0000"}, {" Green ", "#008000"}, {"lime", "#00ff00"},
		{"plum", "#dda0dd"}, {"violet", "#ee82ee"}, {"ReBeccAPurple", "#663399"},
		{"#abc", "#aabbcc"}, {"#AB12ef", "#ab12ef"},
		{"RGB ( 255 , 0 , 128 )", "#ff0080"},
		{"hsv(120, 100, 100)", "#00ff00"}, {"HSB (360,100,100)", "#ff0000"},
		{"hsv(0,0,50)", "#808080"}, {"hsv(0,0,49.999999)", "#7f7f7f"},
		{"hsv(0,0,50.000001)", "#808080"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := textparse.ParseColor(tc.input)
			if err != nil || got.Hex() != tc.hex {
				t.Fatalf("got %s, %v; want %s", got.Hex(), err, tc.hex)
			}
		})
	}
	for _, name := range colorkit.Names() {
		want, _ := colorkit.LookupName(name)
		got, err := textparse.ParseColor(strings.ToUpper(name))
		if err != nil || got != want {
			t.Fatalf("CSS %s: got %v, %v; want %v", name, got, err, want)
		}
	}
}

func TestCanonicalStateCommands(t *testing.T) {
	cases := []struct {
		input string
		hex   string
		led   uint8
		fade  time.Duration
	}{
		{"set red", "#ff0000", 0, 0},
		{"SET led 2 to rgb(1, 2, 3) now", "#010203", 2, 0},
		{"set #abc on led first now", "#aabbcc", 1, 0},
		{"set led both to lime", "#00ff00", 0, 0},
		{"off", "#000000", 0, 0}, {"off led second now", "#000000", 2, 0},
		{"OFF on led all NOW", "#000000", 0, 0},
		{"fade to blue over 2s on led 1", "#0000ff", 1, 2 * time.Second},
		{"fade led 2 to #abc in 0.5s", "#aabbcc", 2, 500 * time.Millisecond},
		{"fade hsv(120,100,100) over 0.0005m", "#00ff00", 0, 30 * time.Millisecond},
		{"fade red over 10.9225m", "#ff0000", 0, 655350 * time.Millisecond},
		{"fade red in 0.010000s", "#ff0000", 0, 10 * time.Millisecond},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := textparse.ParseState(tc.input)
			if err != nil || got.Color.Hex() != tc.hex || got.LED != tc.led || got.FadeTime != tc.fade {
				t.Fatalf("got %+v, %v; want %s LED%d %v", got, err, tc.hex, tc.led, tc.fade)
			}
		})
	}
}

func TestFieldFormPermutations(t *testing.T) {
	fields := []string{"led=second", "color=rgb(1, 2, 3)", "fade=20ms"}
	orders := [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	for _, order := range orders {
		input := fields[order[0]] + " " + fields[order[1]] + " " + fields[order[2]]
		got, err := textparse.ParseState(input)
		if err != nil || got.Color != (colorkit.RGB8{R: 1, G: 2, B: 3}) || got.LED != 2 || got.FadeTime != 20*time.Millisecond {
			t.Fatalf("%s: %+v, %v", input, got, err)
		}
	}
	got, err := textparse.ParseState("FADE = 0ms COLOR = HSB(0,0,50) LED = BOTH")
	if err != nil || got.Color.Hex() != "#808080" || got.LED != 0 || got.FadeTime != 0 {
		t.Fatalf("zero field form: %+v, %v", got, err)
	}
}

func TestRejectedStatesReturnZero(t *testing.T) {
	inputs := []string{
		"", "  ", "do not turn led 1 red now", "blink led 1 red 1s", "rainbow",
		"set red blue", "set red #fff", "set red please", "set red not now",
		"set red hold 3s", "set led 1 red for 3s", "set red then off",
		"set red over 2s", "off over 2s", "set on", "set off", "set @mint", "set random",
		"set led 1 red on led 2", "set led 1 led 2 red", "set red on led 1 on led 1",
		"set red now now", "set red now on led 1", "off on", "off led top", "off led 3",
		"fade red", "fade red now", "fade red over 0ms", "fade red over -1s",
		"fade red over +1s", "fade red over 1e2ms", "fade red over 1.0000001s",
		"fade red over 15ms", "fade red over 0.015s", "fade red over 655360ms",
		"fade red over 999999999999999999999999999m", "fade red over 1s now",
		"fade red over 1s over 1s", "fade red over 1s in 1000ms",
		"set rgb(1,2,3,4)", "set rgb(256,0,0)", "set hsb(0,101,100)",
		"led=1 color=red", "color=red fade=0ms", "led=1 fade=0ms",
		"led=1 color=red fade=0ms led=1", "led=1 color=red fade=0ms extra=1",
		"led=1 color=red fade=0ms set red", "set led=1 color=red fade=0ms",
		"led=1 color=red fade=-0ms", "led=1 color=red fade=1ms",
		"led=1 color= fade=0ms", "led= color=red fade=0ms", "led=1 color=red fade=",
		"set red\noff", "set red\r", "set red\u2028off", "set \xff",
		strings.Repeat("x", 4097), "#ff0000L1T0",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			got, err := textparse.ParseState(input)
			if err == nil || got != (textparse.State{}) {
				t.Fatalf("got executable state %+v, %v", got, err)
			}
			var detail *textparse.Error
			if !errors.As(err, &detail) || detail.Code == "" || detail.Message == "" || detail.Hint == "" || detail.Line != 1 || detail.Column < 1 || detail.Start < 0 || detail.End < detail.Start || detail.End > len(input) {
				t.Fatalf("missing or invalid diagnostic: %#v", err)
			}
		})
	}
}

func TestRejectedColorsAndCoreErrors(t *testing.T) {
	for _, input := range []string{"", "on", "off", "@mint", "random", "red blue", "#abc trailing", "rgb(1.5,2,3)", "rgb(-1,2,3)", "rgb(1 2 3)", "rgb(1,2)", "hsv(1e2,100,100)", "hsv(01,100,100)", "hsb(0,100,100.0000001)", "transparent", "currentColor", "set red"} {
		got, err := textparse.ParseColor(input)
		if err == nil || got != (colorkit.RGB8{}) {
			t.Fatalf("%q: partial/successful color %v, %v", input, got, err)
		}
	}
	for _, input := range []string{"rgb(256,0,0)", "hsv(361,100,100)", "hsb(0,101,100)"} {
		_, err := textparse.ParseColor(input)
		if !errors.Is(err, colorkit.ErrRange) {
			t.Fatalf("%q lost core range cause: %v", input, err)
		}
	}
	_, err := textparse.ParseColor("unknown-color")
	if !errors.Is(err, colorkit.ErrSyntax) {
		t.Fatalf("lost core syntax cause: %v", err)
	}
}

func TestOriginalByteAndRunePositions(t *testing.T) {
	input := "\u2003SeT red 🌈"
	_, err := textparse.ParseState(input)
	var detail *textparse.Error
	if !errors.As(err, &detail) || detail.Start != 11 || detail.End != 15 || detail.Line != 1 || detail.Column != 10 || input[detail.Start:detail.End] != "🌈" {
		t.Fatalf("wrong original UTF-8 location: %#v", detail)
	}
	_, err = textparse.ParseState("set 红色 now")
	if !errors.As(err, &detail) || detail.Start != 4 || detail.End != 10 || detail.Column != 5 {
		t.Fatalf("wrong multibyte color location: %#v", detail)
	}
	_, err = textparse.ParseColor("hsv(0, 101, 100)")
	if !errors.As(err, &detail) || detail.Field != "color.s" || detail.Start != 7 || detail.End != 10 {
		t.Fatalf("wrong component location: %#v", detail)
	}
	_, err = textparse.ParseState("fade blue over")
	if !errors.As(err, &detail) || detail.Start != 14 || detail.End != 14 || detail.Column != 15 {
		t.Fatalf("wrong missing-value location: %#v", detail)
	}
}

func TestValidateFadeProtocolBoundary(t *testing.T) {
	for _, d := range []time.Duration{0, 10 * time.Millisecond, 655350 * time.Millisecond} {
		if err := textparse.ValidateFade(d); err != nil {
			t.Fatalf("%v: %v", d, err)
		}
	}
	for _, d := range []time.Duration{-1, time.Nanosecond, 15 * time.Millisecond, 655360 * time.Millisecond, time.Duration(1<<63 - 1)} {
		if err := textparse.ValidateFade(d); err == nil {
			t.Fatalf("accepted unrepresentable duration %v", d)
		}
	}
}

func TestExactInputBoundAndConcurrentParsing(t *testing.T) {
	input := "set red" + strings.Repeat(" ", 4096-len("set red"))
	if _, err := textparse.ParseState(input); err != nil {
		t.Fatalf("rejected bounded input: %v", err)
	}
	done := make(chan error, 12)
	for i := 0; i < 12; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				if _, err := textparse.ParseState("fade hsv(359.999999,99.999999,50) in 20ms on led second"); err != nil {
					done <- err
					return
				}
			}
			done <- nil
		}()
	}
	for i := 0; i < 12; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
}

func TestStateCommentsAfterFullInputValidation(t *testing.T) {
	for _, input := range []string{"set red // comment", "set red\t//comment", "led=0 color=red fade=0ms // saved state"} {
		got, err := textparse.ParseState(input)
		if err != nil || got.Color.Hex() != "#ff0000" || got.FadeTime != 0 {
			t.Fatalf("%q: %+v, %v", input, got, err)
		}
	}
	for _, input := range []string{
		"set red\n// comment", "set red //\ncomment", "set red //\xff",
		"set red //" + strings.Repeat("x", 4096), "//comment", " //comment",
		"set red//comment", "set rgb(1,//comment", "set rgb(1, //comment",
	} {
		got, err := textparse.ParseState(input)
		if err == nil || got != (textparse.State{}) {
			t.Fatalf("%q: got partial/successful state %+v, %v", input, got, err)
		}
	}
	if _, err := textparse.ParseColor("red // comment"); err == nil {
		t.Fatal("ParseColor accepted a comment")
	}
	_, err := textparse.ParseState("//comment")
	var detail *textparse.Error
	if !errors.As(err, &detail) || detail.Code != "missing_field" || detail.Field != "action" || detail.Start != 0 {
		t.Fatalf("comment-only diagnostic: %#v", detail)
	}
	input := "\u2003fade red over // missing duration"
	_, err = textparse.ParseState(input)
	if !errors.As(err, &detail) || detail.Start != strings.Index(input, "//") || detail.End != detail.Start || detail.Column != 16 {
		t.Fatalf("comment boundary lost original location: %#v", detail)
	}
}
