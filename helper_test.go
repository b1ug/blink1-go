package blink1_test

import (
	"encoding/json"
	"fmt"
	"image/color"
	"reflect"
	"testing"
	"time"

	b1 "github.com/b1ug/blink1-go"
)

func TestPreload(t *testing.T) {
	st := time.Now()
	b1.Preload()
	ep := time.Since(st)
	if ep > 100*time.Millisecond {
		t.Errorf("Preload() took too long: %v", ep)
	}
	t.Logf("Preload() took %v", ep)
}

func TestIsRunningOnSupportedOS(t *testing.T) {
	want := true
	if got := b1.IsRunningOnSupportedOS(); got != want {
		t.Errorf("IsRunningOnSupportedOS() = %v, want %v", got, want)
	}
}

func TestStringer(t *testing.T) {
	tests := []struct {
		typ fmt.Stringer
		exp string
	}{
		// as baseline
		{
			typ: time.Millisecond,
			exp: "1ms",
		},
		// for LEDIndex
		{
			typ: b1.LED1,
			exp: "LED 1",
		},
		{
			typ: b1.LED2,
			exp: "LED 2",
		},
		{
			typ: b1.LEDAll,
			exp: "All LED",
		},
		// device things
		{
			typ: b1.DevicePatternState{IsPlaying: true, CurrentPos: 1, LoopStartPos: 2, LoopEndPos: 3, RepeatTimes: 4},
			exp: "▶️{playing=true cur=1 loop=[2,3) left=4}",
		},
		{
			typ: b1.DevicePatternState{IsPlaying: false, CurrentPos: 10, LoopStartPos: 20, LoopEndPos: 30, RepeatTimes: 40},
			exp: "⏸{playing=false cur=10 loop=[20,30) left=40}",
		},
		{
			typ: b1.DeviceLightState{R: 1, G: 2, B: 3, LED: b1.LED1, FadeTimeMsec: 4},
			exp: "🎨{color=#010203 led=1 fade=4ms}",
		},
		{
			typ: b1.DeviceLightState{R: 10, G: 20, B: 30, LED: b1.LEDAll, FadeTimeMsec: 1000},
			exp: "🎨{color=#0A141E led=0 fade=1000ms}",
		},
		// controller things
		{
			typ: b1.PatternState{IsPlaying: true, CurrentPosition: 1, StartPosition: 2, EndPosition: 3, RepeatTimes: 4},
			exp: "▶️(playing=true cur=1 loop=[2,3) left=4)",
		},
		{
			typ: b1.PatternState{IsPlaying: false, CurrentPosition: 10, StartPosition: 20, EndPosition: 30, RepeatTimes: 40},
			exp: "⏸(playing=false cur=10 loop=[20,30) left=40)",
		},
		{
			typ: b1.LightState{Color: color.RGBA{R: 1, G: 2, B: 3, A: 0xff}, LED: b1.LED1, FadeTime: time.Millisecond},
			exp: "🎨(color=#010203 led=1 fade=1ms)",
		},
		{
			typ: b1.LightState{Color: color.RGBA{R: 10, G: 20, B: 30, A: 0xff}, LED: b1.LEDAll, FadeTime: time.Second},
			exp: "🎨(color=#0A141E led=0 fade=1s)",
		},
		// pattern things
		{
			typ: b1.Pattern{
				StartPosition: 0,
				EndPosition:   10,
				RepeatTimes:   1,
				Sequence: []b1.LightState{
					{
						Color:    color.RGBA{255, 0, 0, 255},
						LED:      b1.LED1,
						FadeTime: 1 * time.Second,
					},
					{
						Color:    color.RGBA{0, 255, 0, 255},
						LED:      b1.LED2,
						FadeTime: 2 * time.Second,
					},
				},
			},
			exp: "🎼(loop=[0,10] repeat=1 seq=2)",
		},
		{
			typ: b1.Pattern{
				StartPosition: 10,
				EndPosition:   20,
				RepeatTimes:   0,
			},
			exp: "🎼(loop=[10,20] repeat=∞ seq=0)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.exp, func(t *testing.T) {
			if got := tc.typ.String(); got != tc.exp {
				t.Errorf("Stringer.String(%v) = %v, want %v", tc.typ, got, tc.exp)
			}
		})
	}
}

func TestSerializeLightState(t *testing.T) {
	l1 := b1.LightState{Color: b1.ColorRed, LED: b1.LED1, FadeTime: 256 * time.Millisecond}

	// text encode
	t1, err := l1.MarshalText()
	if err != nil {
		t.Errorf("%T.MarshalText() got error = %v, want nil", l1, err)
	}
	if r := "#FF0000L1T256"; string(t1) != r {
		t.Errorf("%T.MarshalText() got result = %v, want %v", l1, string(t1), r)
	}
	t.Logf("%v %T.MarshalText() = %v", l1, l1, string(t1))

	// json decode
	j1, err := json.Marshal(l1)
	if err != nil {
		t.Errorf("json.Marshal(%v) got error = %v, want nil", l1, err)
	}
	if r := `"#FF0000L1T256"`; string(j1) != r {
		t.Errorf("json.Marshal(%v) got result = %v, want %v", l1, string(j1), r)
	}
	t.Logf("%v json.Marshal() = %v", l1, string(j1))

	// text decode
	var l2 b1.LightState
	if err := l2.UnmarshalText(t1); err != nil {
		t.Errorf("%T.UnmarshalText(%v) got error = %v, want nil", l2, string(t1), err)
	}
	if l2 != l1 {
		t.Errorf("%T.UnmarshalText(%v) got result = %v, want %v", l2, string(t1), l2, l1)
	}
	t.Logf("%T.UnmarshalText(%v) = %v", l2, string(t1), l2)

	// json encode
	var l3 b1.LightState
	if err := json.Unmarshal(j1, &l3); err != nil {
		t.Errorf("json.Unmarshal(%v) got error = %v, want nil", string(j1), err)
	}
	if l3 != l1 {
		t.Errorf("json.Unmarshal(%v) got result = %v, want %v", string(j1), l3, l1)
	}
	t.Logf("json.Unmarshal(%v) = %v", string(j1), l3)

	// invalid text for decode
	var l4 b1.LightState
	if err := l4.UnmarshalText([]byte("invalid")); err == nil {
		t.Errorf("%T.UnmarshalText(%v) got error = %v, want nil", l4, "invalid", err)
	}
	if err := l4.UnmarshalText([]byte("")); err == nil {
		t.Errorf("%T.UnmarshalText(%v) got error = %v, want nil", l4, "", err)
	}
	if err := l4.UnmarshalText([]byte("#FF0000L1TX256")); err == nil {
		t.Errorf("%T.UnmarshalText(%v) got error = %v, want nil", l4, "#FF0000L1T256;", err)
	}
	if err := json.Unmarshal([]byte(`"#FF0000M1T256"`), &l4); err == nil {
		t.Errorf("json.Unmarshal(%v) got error = %v, want nil", `"#FF0000M1T256"`, err)
	}
}

func TestSerializeStateSequence(t *testing.T) {
	l1 := b1.LightState{Color: b1.ColorRed, LED: b1.LED1, FadeTime: 256 * time.Millisecond}
	l2 := b1.LightState{Color: b1.ColorGreen, LED: b1.LED2, FadeTime: 512 * time.Millisecond}
	l3 := b1.LightState{Color: b1.ColorBlue, LED: b1.LEDAll, FadeTime: 1024 * time.Millisecond}
	s1 := b1.StateSequence{l1, l2, l3}

	// text encode
	t1, err := s1.MarshalText()
	if err != nil {
		t.Errorf("%T.MarshalText() got error = %v, want nil", s1, err)
	}
	if r := "#FF0000L1T256;#00FF00L2T512;#0000FFL0T1024"; string(t1) != r {
		t.Errorf("%T.MarshalText() got result = %v, want %v", s1, string(t1), r)
	}
	t.Logf("%v %T.MarshalText() = %v", s1, s1, string(t1))

	// json decode
	j1, err := json.Marshal(s1)
	if err != nil {
		t.Errorf("json.Marshal(%v) got error = %v, want nil", s1, err)
	}
	if r := `"#FF0000L1T256;#00FF00L2T512;#0000FFL0T1024"`; string(j1) != r {
		t.Errorf("json.Marshal(%v) got result = %v, want %v", s1, string(j1), r)
	}
	t.Logf("%v json.Marshal() = %v", s1, string(j1))

	// text decode
	var s2 b1.StateSequence
	if err := s2.UnmarshalText(t1); err != nil {
		t.Errorf("%T.UnmarshalText(%v) got error = %v, want nil", s2, string(t1), err)
	}
	if !reflect.DeepEqual(s2, s1) {
		t.Errorf("%T.UnmarshalText(%v) got result = %v, want %v", s2, string(t1), s2, s1)
	}
	t.Logf("%T.UnmarshalText(%v) = %v", s2, string(t1), s2)

	// json encode
	var s3 b1.StateSequence
	if err := json.Unmarshal(j1, &s3); err != nil {
		t.Errorf("json.Unmarshal(%v) got error = %v, want nil", string(j1), err)
	}
	if !reflect.DeepEqual(s3, s1) {
		t.Errorf("json.Unmarshal(%v) got result = %v, want %v", string(j1), s3, s1)
	}
	t.Logf("json.Unmarshal(%v) = %v", string(j1), s3)

	// error cases
	var s4 b1.StateSequence
	if err := s4.UnmarshalText([]byte("#FF0000L1T256;FF0000L1T256")); err == nil {
		t.Errorf("%T.UnmarshalText(%v) got error = %v, want nil", s4, "#FF0000L1T256 #FF0000L1T256", err)
	}

	// empty sequence
	var sb b1.StateSequence
	t2, err := sb.MarshalText()
	if err != nil {
		t.Errorf("%T.MarshalText() got error = %v, want nil", sb, err)
	}
	if r := ""; string(t2) != r {
		t.Errorf("%T.MarshalText() got result = %v, want %v", sb, string(t2), r)
	}
	t.Logf("%v %T.MarshalText() = %v", sb, sb, string(t2))

	if err := sb.UnmarshalText(t2); err != nil {
		t.Errorf("%T.UnmarshalText(%v) got error = %v, want nil", sb, string(t2), err)
	}
	if !reflect.DeepEqual(sb, b1.StateSequence{}) {
		t.Errorf("%T.UnmarshalText(%v) got result = %v, want %v", sb, string(t2), sb, b1.StateSequence{})
	}
	t.Logf("%T.UnmarshalText(%v) = %v", sb, string(t2), sb)

	// one element sequence
	sc := b1.StateSequence{l1}
	t3, err := sc.MarshalText()
	if err != nil {
		t.Errorf("%T.MarshalText() got error = %v, want nil", sc, err)
	}
	if r := "#FF0000L1T256"; string(t3) != r {
		t.Errorf("%T.MarshalText() got result = %v, want %v", sc, string(t3), r)
	}
	t.Logf("%v %T.MarshalText() = %v", sc, sc, string(t3))
}

func TestHSBToRGB(t *testing.T) {
	type hsb struct {
		hue, saturation, brightness float64
	}
	type rgb struct {
		r, g, b uint8
	}
	tests := []struct {
		name     string
		hsbValue hsb
		rgbValue rgb
	}{
		{"Black", hsb{0, 0, 0}, rgb{0, 0, 0}},
		{"White", hsb{0, 0, 100}, rgb{255, 255, 255}},
		{"Red", hsb{0, 100, 100}, rgb{255, 0, 0}},
		{"Green", hsb{120, 100, 100}, rgb{0, 255, 0}},
		{"Blue", hsb{240, 100, 100}, rgb{0, 0, 255}},
		{"Yellow", hsb{60, 100, 100}, rgb{255, 255, 0}},
		{"Cyan", hsb{180, 100, 100}, rgb{0, 255, 255}},
		{"Magenta", hsb{300, 100, 100}, rgb{255, 0, 255}},
		{"Grey", hsb{0, 0, 50}, rgb{128, 128, 128}},
		{"Violet", hsb{270, 100, 100}, rgb{128, 0, 255}},
		{"No Saturation", hsb{270, 0, 50}, rgb{128, 128, 128}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b := b1.HSBToRGB(tc.hsbValue.hue, tc.hsbValue.saturation, tc.hsbValue.brightness)
			if r != tc.rgbValue.r || g != tc.rgbValue.g || b != tc.rgbValue.b {
				t.Errorf("Expected: %v, got: R:%v, G:%v, B:%v", tc.rgbValue, r, g, b)
			}
		})
	}
}

func TestColorToHex(t *testing.T) {
	tests := []struct {
		col  color.Color
		want string
	}{
		{
			col:  color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
			want: "#FF0000",
		},
		{
			col:  color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
			want: "#AABBCC",
		},
		{
			col:  color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0x0},
			want: "#AABBCC",
		},
		{
			col:  color.RGBA{R: 0x1, G: 0x2, B: 0x3, A: 0xff},
			want: "#010203",
		},
		{
			col:  color.RGBA{R: 0x80, G: 0x0, B: 0x81, A: 0xff},
			want: "#800081",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := b1.ColorToHex(tt.col)
			if got != tt.want {
				t.Errorf("ColorToHex(%v) got = %s, want = %s", tt.col, got, tt.want)
			}
		})
	}
}

func TestHexToColor(t *testing.T) {
	tests := []struct {
		hex     string
		want    color.Color
		wantErr bool
	}{
		{
			hex:  "#000",
			want: color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
		},
		{
			hex:  "#000000",
			want: color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
		},
		{
			hex:  "000",
			want: color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
		},
		{
			hex:  "000000",
			want: color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
		},
		{
			hex:  "#FF1234",
			want: color.RGBA{R: 0xff, G: 0x12, B: 0x34, A: 0xff},
		},
		{
			hex:  "ff1234",
			want: color.RGBA{R: 0xff, G: 0x12, B: 0x34, A: 0xff},
		},
		{
			hex:  "#AABBCC",
			want: color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		},
		{
			hex:  "#aabbcc",
			want: color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		},
		{
			hex:  "#abc",
			want: color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		},
		{
			hex:  "#010203",
			want: color.RGBA{R: 0x1, G: 0x2, B: 0x3, A: 0xff},
		},
		{
			hex:  "#123",
			want: color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff},
		},
		{
			hex:     "#FF1234 ",
			wantErr: true,
		},
		{
			hex:     " #FF1234",
			wantErr: true,
		},
		{
			hex:     "#FF12345",
			wantErr: true,
		},
		{
			hex:     "#12",
			wantErr: true,
		},
		{
			hex:     "#1234",
			wantErr: true,
		},
		{
			hex:     "X123",
			wantErr: true,
		},
		{
			hex:     "#abg",
			wantErr: true,
		},
		{
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			got, err := b1.HexToColor(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexToColor(%v) error = %v, wantErr = %v", tt.hex, err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("HexToColor(%v) got = %v, want = %v", tt.hex, got, tt.want)
			}
		})
	}
}

func TestRGBToColor(t *testing.T) {
	tests := []struct {
		r, g, b uint8
		want    color.Color
	}{
		{
			want: color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
		},
		{
			r:    0xff,
			g:    0x0,
			b:    0x0,
			want: color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
		},
		{
			r:    0xaa,
			g:    0xbb,
			b:    0xcc,
			want: color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		},
		{
			r:    0x1,
			g:    0x2,
			b:    0x3,
			want: color.RGBA{R: 0x1, G: 0x2, B: 0x3, A: 0xff},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.want), func(t *testing.T) {
			got := b1.RGBToColor(tt.r, tt.g, tt.b)
			if got != tt.want {
				t.Errorf("RGBToColor(%v, %v, %v) got = %v, want = %v", tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}

func TestColorToRGB(t *testing.T) {
	tests := []struct {
		col                 color.Color
		wantR, wantG, wantB uint8
	}{
		{
			col:   color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
			wantR: 0x0,
			wantG: 0x0,
			wantB: 0x0,
		},
		{
			col:   color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
			wantR: 0xff,
			wantG: 0x0,
			wantB: 0x0,
		},
		{
			col:   color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
			wantR: 0xaa,
			wantG: 0xbb,
			wantB: 0xcc,
		},
		{
			col:   color.RGBA{R: 0x1, G: 0x2, B: 0x3, A: 0xff},
			wantR: 0x1,
			wantG: 0x2,
			wantB: 0x3,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.col), func(t *testing.T) {
			gotR, gotG, gotB := b1.ColorToRGB(tt.col)
			if gotR != tt.wantR || gotG != tt.wantG || gotB != tt.wantB {
				t.Errorf("ColorToRGB(%v) got = (%v, %v, %v), want = (%v, %v, %v)", tt.col, gotR, gotG, gotB, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		hex                 string
		wantR, wantG, wantB uint8
		wantErr             bool
	}{
		{
			hex:     "#000",
			wantR:   0x0,
			wantG:   0x0,
			wantB:   0x0,
			wantErr: false,
		},
		{
			hex:     "#000000",
			wantR:   0x0,
			wantG:   0x0,
			wantB:   0x0,
			wantErr: false,
		},
		{
			hex:     "000",
			wantR:   0x0,
			wantG:   0x0,
			wantB:   0x0,
			wantErr: false,
		},
		{
			hex:     "000000",
			wantR:   0x0,
			wantG:   0x0,
			wantB:   0x0,
			wantErr: false,
		},
		{
			hex:     "#FF1234",
			wantR:   0xff,
			wantG:   0x12,
			wantB:   0x34,
			wantErr: false,
		},
		{
			hex:     "ff1234",
			wantR:   0xff,
			wantG:   0x12,
			wantB:   0x34,
			wantErr: false,
		},
		{
			hex:     "#AABBCC",
			wantR:   0xaa,
			wantG:   0xbb,
			wantB:   0xcc,
			wantErr: false,
		},
		{
			hex:     "#aabbcc",
			wantR:   0xaa,
			wantG:   0xbb,
			wantB:   0xcc,
			wantErr: false,
		},
		{
			hex:     "hello!",
			wantErr: true,
		},
		{
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			gotR, gotG, gotB, err := b1.HexToRGB(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexToRGB(%v) error = %v, wantErr = %v", tt.hex, err, tt.wantErr)
				return
			}
			if err == nil && (gotR != tt.wantR || gotG != tt.wantG || gotB != tt.wantB) {
				t.Errorf("HexToRGB(%v) got = (%v, %v, %v), want = (%v, %v, %v)", tt.hex, gotR, gotG, gotB, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestRGBToHex(t *testing.T) {
	tests := []struct {
		r, g, b uint8
		want    string
	}{
		{
			want: "#000000",
		},
		{
			r:    0xff,
			g:    0x0,
			b:    0x0,
			want: "#FF0000",
		},
		{
			r:    0xaa,
			g:    0xbb,
			b:    0xcc,
			want: "#AABBCC",
		},
		{
			r:    0x1,
			g:    0x2,
			b:    0x3,
			want: "#010203",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := b1.RGBToHex(tt.r, tt.g, tt.b)
			if got != tt.want {
				t.Errorf("RGBToHex(%v, %v, %v) got = %v, want = %v", tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}
