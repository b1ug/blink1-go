package blink1_test

import (
	"errors"
	"image/color"
	"strings"
	"testing"
	"time"

	b1 "github.com/b1ug/blink1-go"
	"github.com/b1ug/colorkit"
)

func TestHexCoreLegacyFormatsAndConcreteColor(t *testing.T) {
	cases := []struct {
		input string
		want  color.RGBA
	}{
		{"abc", color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}},
		{"#AbC", color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}},
		{"12abEF", color.RGBA{R: 0x12, G: 0xab, B: 0xef, A: 0xff}},
		{"#12abEF", color.RGBA{R: 0x12, G: 0xab, B: 0xef, A: 0xff}},
		{" \t\r\n\v\f#AbC \t", color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}},
		{" \tAbC\r\n", color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}},
		{strings.Repeat(" ", 61) + "abc", color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}},
		{strings.Repeat(" ", 60) + "#abc", color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}},
	}
	for _, x := range cases {
		t.Run(x.input, func(t *testing.T) {
			got, err := b1.HexToColor(x.input)
			if err != nil || got != x.want {
				t.Fatalf("HexToColor(%q)=%#v,%v, want %#v", x.input, got, err, x.want)
			}
			if _, ok := got.(color.RGBA); !ok {
				t.Fatalf("returned %T, want color.RGBA", got)
			}
			r, g, b, err := b1.HexToRGB(x.input)
			if err != nil || r != x.want.R || g != x.want.G || b != x.want.B {
				t.Fatal(r, g, b, err)
			}
		})
	}
	for _, x := range []struct {
		rgb  [3]uint8
		want string
	}{{[3]uint8{0, 0, 0}, "#000000"}, {[3]uint8{0x12, 0xab, 0xef}, "#12ABEF"}, {[3]uint8{0xff, 0xff, 0xff}, "#FFFFFF"}} {
		if got := b1.RGBToHex(x.rgb[0], x.rgb[1], x.rgb[2]); got != x.want {
			t.Fatal(got, x.want)
		}
	}
}

func TestHexCoreRejectsScannerWhitespaceAndPartialLiterals(t *testing.T) {
	inputs := []string{
		"12 345", "#12 345", "12345 ", "#12345 ", "1 2 3 ", "#1 2 3 ", "F F F ",
		"12+345", "#12+345", "", "#", "ab", "#ab", "#abcd", "#aabbccdd", "aabbccx",
		"#ggg", "red", "rgb(1,2,3)", "#abc trailing", "abc\x00", "\u00a0abc", "#abc\u2003",
		strings.Repeat(" ", 62) + "abc", strings.Repeat(" ", 61) + "#abc",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			got, err := b1.HexToColor(input)
			if got != nil || !errors.Is(err, colorkit.ErrSyntax) {
				t.Fatalf("HexToColor(%q)=%#v,%v", input, got, err)
			}
			var detail *colorkit.Error
			if !errors.As(err, &detail) || detail == nil || detail.Op == "" {
				t.Fatalf("missing core error: %v", err)
			}
			r, g, b, err := b1.HexToRGB(input)
			if r != 0 || g != 0 || b != 0 || !errors.Is(err, colorkit.ErrSyntax) {
				t.Fatal("partial RGB result", r, g, b, err)
			}
		})
	}
}

func TestHexCoreParserHexRoutesKeepSDKRepresentations(t *testing.T) {
	for _, x := range []struct {
		literal string
		want    color.RGBA
	}{{"#AbC", color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}}, {"#12abEF", color.RGBA{R: 0x12, G: 0xab, B: 0xef, A: 0xff}}} {
		got, err := b1.ParseColor(x.literal)
		if err != nil || got != x.want {
			t.Fatal(x.literal, got, err)
		}
		st, err := b1.ParseStateQuery("fade led 2 to " + x.literal + " over 20ms")
		if err != nil || st.Color != x.want || st.LED != b1.LED2 || st.FadeTime != 20*time.Millisecond {
			t.Fatal(x.literal, st, err)
		}
		text, err := st.MarshalText()
		wantText := b1.RGBToHex(x.want.R, x.want.G, x.want.B) + "L2T20"
		if err != nil || string(text) != wantText {
			t.Fatal(string(text), err, wantText)
		}
	}
}
