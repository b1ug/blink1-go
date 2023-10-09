package blink1_test

import (
	"image/color"
	"reflect"
	"testing"
	"time"

	"github.com/b1ug/blink1-go"
)

func BenchmarkParseStateQuery_Simple(b *testing.B) {
	q := `(led:2, color:pink, time:500ms)`
	blink1.ParseStateQuery(q)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blink1.ParseStateQuery(q)
	}
}

func BenchmarkParseStateQuery_Complex(b *testing.B) {
	q := `slowly change all leds to color #add8e6 in 4.5 seconds`
	blink1.ParseStateQuery(q)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blink1.ParseStateQuery(q)
	}
}

func TestParseStateQuery(t *testing.T) {
	tests := []struct {
		query   string
		want    blink1.LightState
		wantErr bool
	}{
		{
			query: "turn off all lights right now",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "set led 1 to color #ff00ff over 2 sec",
			want:  blink1.LightState{Color: blink1.ColorMagenta, LED: blink1.LED1, FadeTime: 2 * time.Second},
		},
		{
			query: "led=0 color=yellow time:500 ms",
			want:  blink1.LightState{Color: blink1.ColorYellow, LED: blink1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "led=1 color=pink fade=10ms",
			want:  blink1.LightState{Color: blink1.ColorPink, LED: blink1.LED1, FadeTime: 10 * time.Millisecond},
		},
		{
			query: "all leds shift to color #ffff00 in no time",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xff, B: 0x0, A: 0xff}, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "all leds transition to black in 2 seconds",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LEDAll, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "change all leds to orange with a fade time of 500 milliseconds",
			want:  blink1.LightState{Color: blink1.ColorOrange, LED: blink1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "change led 1 to lime over 5 seconds",
			want:  blink1.LightState{Color: blink1.ColorLime, LED: blink1.LED1, FadeTime: 5000 * time.Millisecond},
		},
		{
			query: "change light 1 to pink over 5.5secs",
			want:  blink1.LightState{Color: blink1.ColorPink, LED: blink1.LED1, FadeTime: 5500 * time.Millisecond},
		},
		{
			query: "change led 2 to color #ff4500 over 4 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0x45, B: 0x0, A: 0xff}, LED: blink1.LED2, FadeTime: 4000 * time.Millisecond},
		},
		{
			query: "convert led 2 to color #ffc0cb gradually over 3 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xc0, B: 0xcb, A: 0xff}, LED: blink1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "fade all leds off this instant",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "fade top led to red now",
			want:  blink1.LightState{Color: blink1.ColorRed, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "fade bottom led to blue in 2secs",
			want:  blink1.LightState{Color: blink1.ColorBlue, LED: blink1.LED2, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "fade all leds to purple in 1 second",
			want:  blink1.LightState{Color: blink1.ColorPurple, LED: blink1.LEDAll, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "fade led 1 to colour #8b0000 in 1.5 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0x8b, G: 0x0, B: 0x0, A: 0xff}, LED: blink1.LED1, FadeTime: 1500 * time.Millisecond},
		},
		{
			query: "fade led1 to green over 500 milliseconds.",
			want:  blink1.LightState{Color: blink1.ColorGreen, LED: blink1.LED1, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "fade led 1 to red in 2 seconds.",
			want:  blink1.LightState{Color: blink1.ColorRed, LED: blink1.LED1, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "fade led 2 to yellow in 3 seconds",
			want:  blink1.LightState{Color: blink1.ColorYellow, LED: blink1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "gradually change all leds to color #00ff00 in 2.5 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0x0, G: 0xff, B: 0x0, A: 0xff}, LED: blink1.LEDAll, FadeTime: 2500 * time.Millisecond},
		},
		{
			query: "gradually change led 2 to black in 4 seconds",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LED2, FadeTime: 4000 * time.Millisecond},
		},
		{
			query: "immediately change led 2 to color #8b4513",
			want:  blink1.LightState{Color: color.RGBA{R: 0x8b, G: 0x45, B: 0x13, A: 0xff}, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "immediately set all leds to blue",
			want:  blink1.LightState{Color: blink1.ColorBlue, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "immediately set led 2 to color white",
			want:  blink1.LightState{Color: blink1.ColorWhite, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "immediately turn led 1 to white",
			want:  blink1.LightState{Color: blink1.ColorWhite, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "immediately turn off bottom led",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "instantaneously turn all leds to pink",
			want:  blink1.LightState{Color: blink1.ColorPink, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "instantly change top led to cyan",
			want:  blink1.LightState{Color: blink1.ColorCyan, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "instantly turn off all leds ",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "instantly turn off led 2",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "second led changes to orange immediately",
			want:  blink1.LightState{Color: blink1.ColorOrange, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "1st light fades to color #0f2 over 3 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0x0, G: 0xff, B: 0x22, A: 0xff}, LED: blink1.LED1, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "led one instantaneously changes to #ffd700 ",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xd7, B: 0x0, A: 0xff}, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "turns off first led right now.",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "led2 fades to pink over 5 seconds.",
			want:  blink1.LightState{Color: blink1.ColorPink, LED: blink1.LED2, FadeTime: 5000 * time.Millisecond},
		},
		{
			query: "led 2 gradually changes to color #ff0000 in 4 seconds.",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, LED: blink1.LED2, FadeTime: 4000 * time.Millisecond},
		},
		{
			query: "2nd led transitions to green quickly.",
			want:  blink1.LightState{Color: blink1.ColorGreen, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "led 2 transitions to yellow slowly over 3 seconds.",
			want:  blink1.LightState{Color: blink1.ColorYellow, LED: blink1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "led  2 turns blue gradually over 2 seconds.",
			want:  blink1.LightState{Color: blink1.ColorBlue, LED: blink1.LED2, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "led:2, color:green, time:1s",
			want:  blink1.LightState{Color: blink1.ColorGreen, LED: blink1.LED2, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "led:all color:blue fade:500ms",
			want:  blink1.LightState{Color: blink1.ColorBlue, LED: blink1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "make led#2 purple now",
			want:  blink1.LightState{Color: blink1.ColorPurple, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "quickly turn all leds to colour #fa8072 in 1 second",
			want:  blink1.LightState{Color: color.RGBA{R: 0xfa, G: 0x80, B: 0x72, A: 0xff}, LED: blink1.LEDAll, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "red for all in 1s",
			want:  blink1.LightState{Color: blink1.ColorRed, LED: blink1.LEDAll, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "set all led to #00ffff this moment",
			want:  blink1.LightState{Color: color.RGBA{R: 0x0, G: 0xff, B: 0xff, A: 0xff}, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "set all light to blue instantly.",
			want:  blink1.LightState{Color: blink1.ColorBlue, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "set all lights to color #ffeeab immediately.",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xee, B: 0xab, A: 0xff}, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "set all leds to magenta over 2 seconds",
			want:  blink1.LightState{Color: blink1.ColorMagenta, LED: blink1.LEDAll, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "set all leds to red instantly",
			want:  blink1.LightState{Color: blink1.ColorRed, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "set led:1 to blue in 500ms",
			want:  blink1.LightState{Color: blink1.ColorBlue, LED: blink1.LED1, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "set led #1 to green instantly",
			want:  blink1.LightState{Color: blink1.ColorGreen, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "set led  1 to purple right now",
			want:  blink1.LightState{Color: blink1.ColorPurple, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "set led  2 to blue right now",
			want:  blink1.LightState{Color: blink1.ColorBlue, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "set led:2 green over 1 second",
			want:  blink1.LightState{Color: blink1.ColorGreen, LED: blink1.LED2, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "slowly change all leds to color #add8e6 in 4.5 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0xad, G: 0xd8, B: 0xe6, A: 0xff}, LED: blink1.LEDAll, FadeTime: 4500 * time.Millisecond},
		},
		{
			query: "slowly fade all leds to black in 5 seconds",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LEDAll, FadeTime: 5000 * time.Millisecond},
		},
		{
			query: "slowly fade all leds to color #00f over 1.5 seconds.",
			want:  blink1.LightState{Color: color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff}, LED: blink1.LEDAll, FadeTime: 1500 * time.Millisecond},
		},
		{
			query: "transition all leds to white over a period of 500 milliseconds",
			want:  blink1.LightState{Color: blink1.ColorWhite, LED: blink1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "transition led 1 to color #808080 in 2.5 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}, LED: blink1.LED1, FadeTime: 2500 * time.Millisecond},
		},
		{
			query: "turn all leds to #f08080 swiftly over 1 second",
			want:  blink1.LightState{Color: color.RGBA{R: 0xf0, G: 0x80, B: 0x80, A: 0xff}, LED: blink1.LEDAll, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "turn led 1 to color #ffa500 over 3 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xa5, B: 0x0, A: 0xff}, LED: blink1.LED1, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "turn led 2 to #ffa500 over 3 seconds",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xa5, B: 0x0, A: 0xff}, LED: blink1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "turn off all leds this instant",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "turn on all now",
			want:  blink1.LightState{Color: blink1.ColorWhite, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "rgb(255, 0, 255) for both now",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0x0, B: 0xff, A: 0xff}, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "rgb(255,255,0) for all now",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xff, B: 0x0, A: 0xff}, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "hsb(195, 100, 100) for led 1 instantly",
			want:  blink1.LightState{Color: color.RGBA{R: 0x0, G: 0xbf, B: 0xff, A: 0xff}, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "set hsb(315,100,100) for led2 right now",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0x0, B: 0xbf, A: 0xff}, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: "all lights off now",
			want:  blink1.LightState{Color: blink1.ColorBlack, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: "set red for led#2 in 2 secs",
			want:  blink1.LightState{Color: blink1.ColorRed, LED: blink1.LED2, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "make led one blue over 600 msecs",
			want:  blink1.LightState{Color: blink1.ColorBlue, LED: blink1.LED1, FadeTime: 600 * time.Millisecond},
		},
		{
			query: "turn both light green within 800ms",
			want:  blink1.LightState{Color: blink1.ColorGreen, LED: blink1.LEDAll, FadeTime: 800 * time.Millisecond},
		},
		{
			query: "🎨(color=#800080 led=1 fade=0ms)",
			want:  blink1.LightState{Color: color.RGBA{R: 0x80, G: 0x0, B: 0x80, A: 0xff}, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: "🎨(color=#FFA500 led=2 fade=3s)",
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xa5, B: 0x0, A: 0xff}, LED: blink1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: `(led:0, color:yellow, time:500ms)`,
			want:  blink1.LightState{Color: blink1.ColorYellow, LED: blink1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: `(led:2, color:#00ff00, time:2s)`,
			want:  blink1.LightState{Color: blink1.ColorGreen, LED: blink1.LED2, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: `(led=all, color=#00ff00, time=1m)`,
			want:  blink1.LightState{Color: blink1.ColorGreen, LED: blink1.LEDAll, FadeTime: 1 * time.Minute},
		},
		{
			query: `(led:top, color=#faf, time=0)`,
			want:  blink1.LightState{Color: color.RGBA{R: 0xff, G: 0xaa, B: 0xff, A: 0xff}, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query: `(led:btm, color=#abc, time: 0)`,
			want:  blink1.LightState{Color: color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}, LED: blink1.LED2, FadeTime: 0},
		},
		{
			query: `(led:1, color=#abcdef, fade=1s)`,
			want:  blink1.LightState{Color: color.RGBA{R: 0xab, G: 0xcd, B: 0xef, A: 0xff}, LED: blink1.LED1, FadeTime: 1 * time.Second},
		},
		{
			query: `all pink now`,
			want:  blink1.LightState{Color: blink1.ColorPink, LED: blink1.LEDAll, FadeTime: 0},
		},
		{
			query: `led=1 color=yellow time=500ms`,
			want:  blink1.LightState{Color: blink1.ColorYellow, LED: blink1.LED1, FadeTime: 500 * time.Millisecond},
		},
		{
			query: `led=1 color=yellow now`,
			want:  blink1.LightState{Color: blink1.ColorYellow, LED: blink1.LED1, FadeTime: 0},
		},
		{
			query:   `led=1 color=yellow`,
			wantErr: true,
		},
		{
			query:   `color=yellow now`,
			wantErr: true,
		},
		{
			query:   `led color=yellow time=500ms`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := blink1.ParseStateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStateQuery(%q) got error = %v, wantErr %v", tt.query, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseStateQuery(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
