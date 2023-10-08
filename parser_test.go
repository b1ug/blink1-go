package blink1_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/b1ug/blink1-go"
)

func TestParseStateQuery(t *testing.T) {
	tests := []struct {
		query   string
		want    blink1.LightState
		wantErr bool
	}{
		{
			query:   "led=0 color=yellow time:500 ms",
			want:    blink1.LightState{Color: blink1.ColorYellow, LED: blink1.LEDAll, FadeTime: 500 * time.Millisecond},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := blink1.ParseStateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStateQuery(%q) got error = %v, wantErr %v", tt.query, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseStateQuery(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
