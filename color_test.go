package blink1_test

import (
	"fmt"
	"image/color"
	"reflect"
	"testing"

	b1 "github.com/b1ug/blink1-go"
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
	if ns1[0] != "apricot" {
		t.Errorf("GetColorNames() should return apricot as first element, got %v", ns1[0])
	}

	ns1[0] = "foo"
	ns2[0] = "bar"
	ns3 := b1.GetColorNames()
	if ns3[0] != "apricot" {
		t.Errorf("GetColorNames() should not be mutable, got %v", ns3[0])
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
