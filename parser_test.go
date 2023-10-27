package blink1_test

import (
	"image/color"
	"reflect"
	"testing"
	"time"

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
	q := `(led:2, color:pink, time:500ms)`
	b1.ParseStateQuery(q)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b1.ParseStateQuery(q)
	}
}

func BenchmarkParseStateQuery_Complex(b *testing.B) {
	q := `slowly change all leds to color #add8e6 in 4.5 seconds`
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

func TestParseColor(t *testing.T) {
	tests := []struct {
		query   string
		want    color.Color
		wantErr bool
	}{
		{
			query: "#ff0000",
			want:  color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
		},
		{
			query: "#00FF00",
			want:  color.RGBA{R: 0x0, G: 0xff, B: 0x0, A: 0xff},
		},
		{
			query: "#abc",
			want:  color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		},
		{
			query: "on",
			want:  color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		},
		{
			query: "off",
			want:  color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
		},
		{
			query: "red",
			want:  color.RGBA{R: 0xff, G: 0x00, B: 0x0, A: 0xff},
		},
		{
			query: "BLUE",
			want:  color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff},
		},
		{
			query: "rgb(12,34,56)",
			want:  color.RGBA{R: 0x0c, G: 0x22, B: 0x38, A: 0xff},
		},
		{
			query: "hsb(356, 64, 90)",
			want:  color.RGBA{R: 0xe6, G: 0x53, B: 0x5c, A: 0xff},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := b1.ParseColor(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseColor(%q) error = %v, wantErr = %v", tt.query, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got != tt.want {
				t.Errorf("ParseColor(%q) got = %v, want = %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestParseStateQuery(t *testing.T) {
	tests := []struct {
		query   string
		want    b1.LightState
		wantErr bool
	}{
		{
			query: "turn off all lights right now",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "turn on all lights right now",
			want:  b1.LightState{Color: b1.ColorWhite, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "set led 1 to color #ff00ff over 2 sec",
			want:  b1.LightState{Color: b1.ColorMagenta, LED: b1.LED1, FadeTime: 2 * time.Second},
		},
		{
			query: "led=0 color=yellow time:500 ms",
			want:  b1.LightState{Color: b1.ColorYellow, LED: b1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "led=1 color=pink fade=10ms",
			want:  b1.LightState{Color: b1.ColorPink, LED: b1.LED1, FadeTime: 10 * time.Millisecond},
		},
		{
			query: "all leds shift to color #ffff00 in no time",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xff, B: 0x0, A: 0xff}, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "all leds transition to black in 2 seconds",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LEDAll, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "change all leds to orange with a fade time of 500 milliseconds",
			want:  b1.LightState{Color: b1.ColorOrange, LED: b1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "change led 1 to lime over 5 seconds",
			want:  b1.LightState{Color: b1.ColorLime, LED: b1.LED1, FadeTime: 5000 * time.Millisecond},
		},
		{
			query: "change light 1 to pink over 5.5secs",
			want:  b1.LightState{Color: b1.ColorPink, LED: b1.LED1, FadeTime: 5500 * time.Millisecond},
		},
		{
			query: "change led 2 to color #ff4500 over 4 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0x45, B: 0x0, A: 0xff}, LED: b1.LED2, FadeTime: 4000 * time.Millisecond},
		},
		{
			query: "convert led 2 to color #ffc0cb gradually over 3 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xc0, B: 0xcb, A: 0xff}, LED: b1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "fade all leds off this instant",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "fade top led to red now",
			want:  b1.LightState{Color: b1.ColorRed, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "fade bottom led to blue in 2secs",
			want:  b1.LightState{Color: b1.ColorBlue, LED: b1.LED2, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "fade all leds to purple in 1 second",
			want:  b1.LightState{Color: b1.ColorPurple, LED: b1.LEDAll, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "fade led 1 to colour #8b0000 in 1.5 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0x8b, G: 0x0, B: 0x0, A: 0xff}, LED: b1.LED1, FadeTime: 1500 * time.Millisecond},
		},
		{
			query: "fade led1 to green over 500 milliseconds.",
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LED1, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "fade led 1 to red in 2 seconds.",
			want:  b1.LightState{Color: b1.ColorRed, LED: b1.LED1, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "fade led 2 to yellow in 3 seconds",
			want:  b1.LightState{Color: b1.ColorYellow, LED: b1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "gradually change all leds to color #00ff00 in 2.5 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0x0, G: 0xff, B: 0x0, A: 0xff}, LED: b1.LEDAll, FadeTime: 2500 * time.Millisecond},
		},
		{
			query: "gradually change led 2 to black in 4 seconds",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LED2, FadeTime: 4000 * time.Millisecond},
		},
		{
			query: "immediately change led 2 to color #8b4513",
			want:  b1.LightState{Color: color.RGBA{R: 0x8b, G: 0x45, B: 0x13, A: 0xff}, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "immediately set all leds to blue",
			want:  b1.LightState{Color: b1.ColorBlue, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "immediately set led 2 to color white",
			want:  b1.LightState{Color: b1.ColorWhite, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "immediately turn led 1 to white",
			want:  b1.LightState{Color: b1.ColorWhite, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "immediately turn off bottom led",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "instantaneously turn all leds to pink",
			want:  b1.LightState{Color: b1.ColorPink, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "instantly change top led to cyan",
			want:  b1.LightState{Color: b1.ColorCyan, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "instantly turn off all leds ",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "instantly turn off led 2",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "second led changes to orange immediately",
			want:  b1.LightState{Color: b1.ColorOrange, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "1st light fades to color #0f2 over 3 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0x0, G: 0xff, B: 0x22, A: 0xff}, LED: b1.LED1, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "led one instantaneously changes to #ffd700 ",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xd7, B: 0x0, A: 0xff}, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "turns off first led right now.",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "led2 fades to pink over 5 seconds.",
			want:  b1.LightState{Color: b1.ColorPink, LED: b1.LED2, FadeTime: 5000 * time.Millisecond},
		},
		{
			query: "led 2 gradually changes to color #ff0000 in 4 seconds.",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff}, LED: b1.LED2, FadeTime: 4000 * time.Millisecond},
		},
		{
			query: "2nd led transitions to green quickly.",
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "led 2 transitions to yellow slowly over 3 seconds.",
			want:  b1.LightState{Color: b1.ColorYellow, LED: b1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "led  2 turns blue gradually over 2 seconds.",
			want:  b1.LightState{Color: b1.ColorBlue, LED: b1.LED2, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "led:2, color:green, time:1s",
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LED2, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "led:all color:blue fade:500ms",
			want:  b1.LightState{Color: b1.ColorBlue, LED: b1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "make led#2 purple now",
			want:  b1.LightState{Color: b1.ColorPurple, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "quickly turn all leds to colour #fa8072 in 1 second",
			want:  b1.LightState{Color: color.RGBA{R: 0xfa, G: 0x80, B: 0x72, A: 0xff}, LED: b1.LEDAll, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "red for all in 1s",
			want:  b1.LightState{Color: b1.ColorRed, LED: b1.LEDAll, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "set all led to #00ffff this moment",
			want:  b1.LightState{Color: color.RGBA{R: 0x0, G: 0xff, B: 0xff, A: 0xff}, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "set all light to blue instantly.",
			want:  b1.LightState{Color: b1.ColorBlue, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "set all lights to color #ffeeab immediately.",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xee, B: 0xab, A: 0xff}, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "set all leds to magenta over 2 seconds",
			want:  b1.LightState{Color: b1.ColorMagenta, LED: b1.LEDAll, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "set all leds to red instantly",
			want:  b1.LightState{Color: b1.ColorRed, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "set led:1 to blue in 500ms",
			want:  b1.LightState{Color: b1.ColorBlue, LED: b1.LED1, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "set led #1 to green instantly",
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "set led  1 to purple right now",
			want:  b1.LightState{Color: b1.ColorPurple, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "set led  2 to blue right now",
			want:  b1.LightState{Color: b1.ColorBlue, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "set led:2 green over 1 second",
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LED2, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "slowly change all leds to color #add8e6 in 4.5 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0xad, G: 0xd8, B: 0xe6, A: 0xff}, LED: b1.LEDAll, FadeTime: 4500 * time.Millisecond},
		},
		{
			query: "slowly fade all leds to black in 5 seconds",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LEDAll, FadeTime: 5000 * time.Millisecond},
		},
		{
			query: "slowly fade all leds to color #00f over 1.5 seconds.",
			want:  b1.LightState{Color: color.RGBA{R: 0x0, G: 0x0, B: 0xff, A: 0xff}, LED: b1.LEDAll, FadeTime: 1500 * time.Millisecond},
		},
		{
			query: "transition all leds to white over a period of 500 milliseconds",
			want:  b1.LightState{Color: b1.ColorWhite, LED: b1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: "transition led 1 to color #808080 in 2.5 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}, LED: b1.LED1, FadeTime: 2500 * time.Millisecond},
		},
		{
			query: "turn all leds to #f08080 swiftly over 1 second",
			want:  b1.LightState{Color: color.RGBA{R: 0xf0, G: 0x80, B: 0x80, A: 0xff}, LED: b1.LEDAll, FadeTime: 1000 * time.Millisecond},
		},
		{
			query: "turn led 1 to color #ffa500 over 3 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xa5, B: 0x0, A: 0xff}, LED: b1.LED1, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "turn led 2 to #ffa500 over 3 seconds",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xa5, B: 0x0, A: 0xff}, LED: b1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: "turn off all leds this instant",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "turn on all now",
			want:  b1.LightState{Color: b1.ColorWhite, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "rgb(255, 0, 255) for both now",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0x0, B: 0xff, A: 0xff}, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "rgb(255,255,0) for all now",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xff, B: 0x0, A: 0xff}, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "hsb(195, 100, 100) for led 1 instantly",
			want:  b1.LightState{Color: color.RGBA{R: 0x0, G: 0xbf, B: 0xff, A: 0xff}, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "set hsb(315,100,100) for led2 right now",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0x0, B: 0xbf, A: 0xff}, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: "all lights off now",
			want:  b1.LightState{Color: b1.ColorBlack, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: "set red for led#2 in 2 secs",
			want:  b1.LightState{Color: b1.ColorRed, LED: b1.LED2, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: "make led one blue over 600 msecs",
			want:  b1.LightState{Color: b1.ColorBlue, LED: b1.LED1, FadeTime: 600 * time.Millisecond},
		},
		{
			query: "turn both light green within 800ms",
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LEDAll, FadeTime: 800 * time.Millisecond},
		},
		{
			query: "Maintain all LEDs on YELLOW for 3.5 seconds",
			want:  b1.LightState{Color: b1.ColorYellow, LED: b1.LEDAll, FadeTime: 3500 * time.Millisecond},
		},
		{
			query: "Maintain the GREEN on all LEDs for 1 second",
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LEDAll, FadeTime: 1 * time.Second},
		},
		{
			query: "🎨(color=#800080 led=1 fade=0ms)",
			want:  b1.LightState{Color: color.RGBA{R: 0x80, G: 0x0, B: 0x80, A: 0xff}, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: "🎨(color=#FFA500 led=2 fade=3s)",
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xa5, B: 0x0, A: 0xff}, LED: b1.LED2, FadeTime: 3000 * time.Millisecond},
		},
		{
			query: `(led:0, color:yellow, time:500ms)`,
			want:  b1.LightState{Color: b1.ColorYellow, LED: b1.LEDAll, FadeTime: 500 * time.Millisecond},
		},
		{
			query: `(led:2, color:#00ff00, time:2s)`,
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LED2, FadeTime: 2000 * time.Millisecond},
		},
		{
			query: `(led=all, color=#00ff00, time=1m)`,
			want:  b1.LightState{Color: b1.ColorGreen, LED: b1.LEDAll, FadeTime: 1 * time.Minute},
		},
		{
			query: `(led:top, color=#faf, time=0)`,
			want:  b1.LightState{Color: color.RGBA{R: 0xff, G: 0xaa, B: 0xff, A: 0xff}, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: `(led:btm, color=#abc, time: 0)`,
			want:  b1.LightState{Color: color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}, LED: b1.LED2, FadeTime: 0},
		},
		{
			query: `(led:1, color=#abcdef, fade=1s)`,
			want:  b1.LightState{Color: color.RGBA{R: 0xab, G: 0xcd, B: 0xef, A: 0xff}, LED: b1.LED1, FadeTime: 1 * time.Second},
		},
		{
			query: `all pink now`,
			want:  b1.LightState{Color: b1.ColorPink, LED: b1.LEDAll, FadeTime: 0},
		},
		{
			query: `led=1 color=yellow time=500ms`,
			want:  b1.LightState{Color: b1.ColorYellow, LED: b1.LED1, FadeTime: 500 * time.Millisecond},
		},
		{
			query: `led=1 color=yellow now`,
			want:  b1.LightState{Color: b1.ColorYellow, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: `led=1 color=yellow now // led=2 color=blue time=500ms`,
			want:  b1.LightState{Color: b1.ColorYellow, LED: b1.LED1, FadeTime: 0},
		},
		{
			query: `#8000FFL0T1500`,
			want:  b1.LightState{Color: color.RGBA{R: 0x80, G: 0x0, B: 0xff, A: 0xff}, LED: b1.LEDAll, FadeTime: 1500 * time.Millisecond},
		},
		{
			query: `#123456l2t20   `,
			want:  b1.LightState{Color: color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}, LED: b1.LED2, FadeTime: 20 * time.Millisecond},
		},
		{
			query:   `#8000FGL0T1500`,
			wantErr: true,
		},
		{
			query:   `#123456l2x20`,
			wantErr: true,
		},
		{
			query:   `123456l2t20`,
			wantErr: true,
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
			got, err := b1.ParseStateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStateQuery(%q) got error = %v, wantErr = %v", tt.query, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseStateQuery(%q) got = %v, want = %v", tt.query, got, tt.want)
			}
		})
	}
}
