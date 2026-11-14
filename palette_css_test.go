package blink1_test

import (
	"image/color"
	"reflect"
	"sort"
	"testing"

	b1 "github.com/b1ug/blink1-go"
	"github.com/b1ug/colorkit"
)

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
