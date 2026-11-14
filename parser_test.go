package blink1_test

import (
	"testing"

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
