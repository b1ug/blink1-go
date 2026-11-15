package blink1_test

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	b1 "github.com/b1ug/blink1-go"
	"github.com/b1ug/colorkit"
)

func TestGetColorNames(t *testing.T) {
	ns1 := b1.GetColorNames()
	ns2 := b1.GetColorNames()

	if !reflect.DeepEqual(ns1, ns2) {
		t.Errorf("GetColorNames() should be consistent, got %v and %v", ns1, ns2)
	}
	if len(ns1) == 0 {
		t.Errorf("GetColorNames() should return non-empty slice, got %v", ns1)
	}
	if ns1[0] != "aliceblue" {
		t.Errorf("GetColorNames() should return aliceblue as first element, got %v", ns1[0])
	}

	ns1[0] = "foo"
	ns2[0] = "bar"
	ns3 := b1.GetColorNames()
	if ns3[0] != "aliceblue" {
		t.Errorf("GetColorNames() should not be mutable, got %v", ns3[0])
	}

	for _, n := range ns3 {
		// get color by name
		c, ok := b1.GetColorByName(n)
		if !ok {
			t.Errorf("GetColorByName(%q) should return true", n)
		}
		// get name by color as reverse
		n2, ok2 := b1.GetNameByColor(c)
		if !ok2 {
			t.Errorf("GetNameByColor(%v) should return true", c)
		}
		// Aliases may have different names; the reverse result must preserve RGB.
		canonical, ok := b1.GetColorByName(n2)
		if !ok || b1.ColorToHex(canonical) != b1.ColorToHex(c) {
			t.Errorf("GetNameByColor(%v) returned %q with a different color", c, n2)
		}
	}
}

func TestGetColorByName(t *testing.T) {
	tests := []struct {
		query string
		want  color.Color
		found bool
	}{
		{
			query: "red",
			want:  color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
			found: true,
		},
		{
			query: "BLUE",
			want:  color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff},
			found: true,
		},
		{
			query: "Yellow",
			want:  color.RGBA{R: 0xff, G: 0xff, B: 0x0, A: 0xff},
			found: true,
		},
		{
			query: "\tpurple    ",
			want:  color.RGBA{R: 0x80, G: 0x0, B: 0x80, A: 0xff},
			found: true,
		},
		{
			query: "none",
			found: false,
		},
		{
			query: "   ",
			found: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, found := b1.GetColorByName(tt.query)
			if found != tt.found || (got != tt.want && tt.found) {
				t.Errorf("GetColorByName(%q) got = (%v, %t), want = (%v, %t)", tt.query, got, found, tt.want, tt.found)
			}
		})
	}
}

func TestGetNameByColor(t *testing.T) {
	tests := []struct {
		col   color.Color
		want  string
		found bool
	}{
		{
			col:   color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
			want:  "red",
			found: true,
		},
		{
			col:   color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff},
			want:  "blue",
			found: true,
		},
		{
			col:   color.RGBA{R: 0xff, G: 0xff, B: 0x0, A: 0xff},
			want:  "yellow",
			found: true,
		},
		{
			col:   color.RGBA{R: 0x80, G: 0x0, B: 0x80, A: 0xff},
			want:  "purple",
			found: true,
		},
		{
			col:   color.RGBA{R: 0x80, G: 0x0, B: 0x80, A: 0x0},
			want:  "purple",
			found: true,
		},
		{
			col:   color.RGBA{R: 0x1, G: 0x2, B: 0x3, A: 0xff},
			found: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got, found := b1.GetNameByColor(tt.col)
			if found != tt.found || (got != tt.want && tt.found) {
				t.Errorf("GetNameByColor(%v) got = (%s, %t), want = (%s, %t)", tt.col, got, found, tt.want, tt.found)
			}
		})
	}
}

func TestGetNameOrHexByColor(t *testing.T) {
	tests := []struct {
		col  color.Color
		want string
	}{
		{
			col:  color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
			want: "red",
		},
		{
			col:  color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff},
			want: "blue",
		},
		{
			col:  color.RGBA{R: 0xff, G: 0xff, B: 0x0, A: 0xff},
			want: "yellow",
		},
		{
			col:  color.RGBA{R: 0x80, G: 0x0, B: 0x80, A: 0xff},
			want: "purple",
		},
		{
			col:  color.RGBA{R: 0x80, G: 0x0, B: 0x80, A: 0x0},
			want: "purple",
		},
		{
			col:  color.RGBA{R: 0x80, G: 0x0, B: 0x81, A: 0xff},
			want: "#800081",
		},
		{
			col:  color.RGBA{R: 0x1, G: 0x2, B: 0x3, A: 0xff},
			want: "#010203",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := b1.GetNameOrHexByColor(tt.col)
			if got != tt.want {
				t.Errorf("GetNameOrHexByColor(%v) got = %s, want = %s", tt.col, got, tt.want)
			}
		})
	}
}

func TestRandomColor(t *testing.T) {
	times := 100
	colors := make([]color.Color, times)
	for i := 0; i < times; i++ {
		colors[i] = b1.RandomColor()
	}
	counts := make(map[string]int)
	for _, c := range colors {
		r, g, b, _ := c.RGBA()
		s := fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
		counts[s]++
	}
	if lc := len(counts); lc <= int(float64(times)*0.9) {
		t.Errorf("RandomColor(*) = %v, want different colors", lc)
	}
}

func TestGammaColor(t *testing.T) {
	for i := 0; i < 256; i++ {
		cl0 := b1.RGBToColor(uint8(i), uint8(i), uint8(i))
		cl1 := b1.DecodeGammaColor(cl0)
		cl2 := b1.EncodeGammaColor(cl1)
		cl3 := b1.DecodeGammaColor(cl2)
		if cl3 != cl1 {
			t.Errorf("Decode(%v) got = %v, Decode(Encode(Decode(%v))) got = %v, different", cl0, cl1, cl0, cl3)
		}
	}
}

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

func TestCSSPaletteNamesAndSDKRepresentation(t *testing.T) {
	names := b1.GetColorNames()
	if len(names) != 148 || !sort.StringsAreSorted(names) || !reflect.DeepEqual(names, colorkit.Names()) {
		t.Fatal("SDK names must match the complete sorted CSS palette")
	}
	for _, name := range names {
		want, ok := colorkit.LookupName(name)
		got, found := b1.GetColorByName(name)
		rgba, concrete := got.(color.RGBA)
		if !ok || !found || !concrete || rgba != (color.RGBA{R: want.R, G: want.G, B: want.B, A: 0xff}) {
			t.Fatalf("%s: got %#v/%t, want opaque color.RGBA", name, got, found)
		}
	}
	names[0] = "local-only"
	if fresh := b1.GetColorNames(); fresh[0] != "aliceblue" || !sort.StringsAreSorted(fresh) {
		t.Fatal("GetColorNames exposed mutable storage")
	}
	if got, ok := b1.GetColorByName(" \tReBeccAPurPle\r\n"); !ok || b1.ColorToHex(got) != "#663399" {
		t.Fatal(got, ok)
	}
	for _, input := range []string{"#abc", "transparent", "currentColor", "\u00a0red"} {
		if got, ok := b1.GetColorByName(input); ok || got != nil {
			t.Fatal(input, got, ok)
		}
	}
}

func TestCSSPaletteReverseAliasPreferences(t *testing.T) {
	groups := []struct {
		a, b, preferred string
	}{
		{"aqua", "cyan", "cyan"},
		{"fuchsia", "magenta", "magenta"},
		{"gray", "grey", "gray"},
		{"darkgray", "darkgrey", "darkgray"},
		{"darkslategray", "darkslategrey", "darkslategray"},
		{"dimgray", "dimgrey", "dimgray"},
		{"lightgray", "lightgrey", "lightgray"},
		{"lightslategray", "lightslategrey", "lightslategray"},
		{"slategray", "slategrey", "slategray"},
	}
	for _, group := range groups {
		for _, name := range []string{group.a, group.b} {
			c, ok := b1.GetColorByName(name)
			if !ok {
				t.Fatal(name)
			}
			for i := 0; i < 4; i++ {
				got, found := b1.GetNameByColor(c)
				if !found || got != group.preferred || b1.GetNameOrHexByColor(c) != group.preferred {
					t.Fatalf("%s: reverse %q/%t, want %q", name, got, found, group.preferred)
				}
			}
		}
	}
	// Keep the SDK's existing alpha-ignoring and high-eight-bit behavior.
	for _, x := range []struct {
		input color.Color
		name  string
		found bool
	}{
		{color.RGBA{R: 0x80, B: 0x80, A: 0}, "purple", true},
		{color.RGBA64{R: 0x00ff, A: 0xffff}, "black", true},
		{color.RGBA64{R: 0x12ff, G: 0x34ff, B: 0x56ff, A: 0xffff}, "#123456", false},
		{color.RGBA{R: 1, G: 2, B: 3, A: 0xff}, "#010203", false},
	} {
		got, found := b1.GetNameByColor(x.input)
		if got != x.name || found != x.found {
			t.Fatal(x.input, got, found, x.name, x.found)
		}
	}
}

func TestCSSPaletteFourStandardValues(t *testing.T) {
	for _, x := range []struct {
		name string
		c    color.RGBA
		hex  string
	}{
		{"green", b1.ColorGreen, "#008000"},
		{"lime", b1.ColorLime, "#00FF00"},
		{"plum", b1.ColorPlum, "#DDA0DD"},
		{"violet", b1.ColorViolet, "#EE82EE"},
	} {
		got, ok := b1.GetColorByName(x.name)
		if !ok || got != x.c || b1.ColorToHex(x.c) != x.hex {
			t.Fatal(x.name, got, ok, x.c, x.hex)
		}
	}
}

func TestLegacyColorConstantsRemainDirectValues(t *testing.T) {
	for _, x := range []struct {
		name string
		c    color.RGBA
		hex  string
	}{
		{"apricot", b1.ColorApricot, "#FBCEB1"},
		{"bronze", b1.ColorBronze, "#CD7F32"},
		{"mint", b1.ColorMint, "#16982B"},
		{"peach", b1.ColorPeach, "#FFE5B4"},
		{"scarlet", b1.ColorScarlet, "#FF2400"},
	} {
		state := b1.NewLightState(x.c, 0, b1.LEDAll)
		if state.Color != x.c || b1.ColorToHex(state.Color) != x.hex {
			t.Fatal("historical Go value changed", x.name, state.Color)
		}
		if got, ok := b1.GetColorByName(x.name); ok || got != nil {
			t.Fatal("non-CSS bare name accepted", x.name, got, ok)
		}
	}
}

func TestRainbowArtAnchorsAreIndependentOfCSSNames(t *testing.T) {
	want := []string{"#FF0000", "#FFA500", "#FFFF00", "#00FF00", "#0000FF", "#4B0082", "#8F00FF"}
	if len(b1.RainbowColors) != len(want) {
		t.Fatalf("rainbow has %d colors, want seven without a repeated endpoint", len(b1.RainbowColors))
	}
	for i, c := range b1.RainbowColors {
		rgba, ok := c.(color.RGBA)
		if !ok || rgba.A != 0xff || b1.ColorToHex(c) != want[i] {
			t.Fatal(i, c, want[i])
		}
	}
	if b1.RainbowColors[3] == b1.ColorGreen || b1.RainbowColors[6] == b1.ColorViolet {
		t.Fatal("artistic anchors must not follow CSS green/violet values")
	}
}

// colorkit values already implement color.Color; pass them to the SDK directly.
func ExampleNewLightState_colorkit() {
	light, err := colorkit.Parse("#abc")
	if err != nil {
		panic(err)
	}
	state := b1.NewLightState(light, 250*time.Millisecond, b1.LED1)
	fmt.Println(state)

	// Output:
	// 🎨(color=#AABBCC led=1 fade=250ms)
}

// HSV validates and converts once, then the resulting value goes straight into
// the SDK. H is in degrees; S and V are percentages.
func ExampleNewLightState_hsv() {
	light, err := (colorkit.HSV{H: 210, S: 75, V: 90}).RGB8()
	if err != nil {
		panic(err)
	}
	state := b1.NewLightState(light, time.Second, b1.LEDAll)
	fmt.Println(state)

	// Output:
	// 🎨(color=#398FE6 led=0 fade=1s)
}

// The existing SDK helper is a single-call option for hexadecimal input.
func ExampleHexToColor_lightState() {
	light, err := b1.HexToColor("abc")
	if err != nil {
		panic(err)
	}
	state := b1.NewLightState(light, 500*time.Millisecond, b1.LED2)
	fmt.Println(state)

	// Output:
	// 🎨(color=#AABBCC led=2 fade=500ms)
}

// This example requires connected hardware. It is compiled, but has no Output
// comment, so go test does not open a controller or execute PlayColor.
func ExampleController_PlayColor_colorkit() {
	light, err := colorkit.Parse("rebeccapurple")
	if err != nil {
		panic(err)
	}
	controller, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer controller.Close()

	if err := controller.PlayColor(light); err != nil {
		panic(err)
	}
}
