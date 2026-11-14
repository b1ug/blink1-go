package blink1

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

// transactionHID exercises query failures without enumerating or opening USB.
type transactionHID struct {
	writeErr  error
	readErr   error
	response  []byte
	readSize  int
	writes    [][]byte
	readInput []byte
	reads     int
}

func (h *transactionHID) Close() {}
func (h *transactionHID) Write(data []byte) error {
	return h.WriteFeature(data)
}
func (h *transactionHID) WriteFeature(data []byte) error {
	h.writes = append(h.writes, append([]byte(nil), data...))
	return h.writeErr
}
func (h *transactionHID) Read(data []byte) (int, error) { return h.ReadFeature(data) }
func (h *transactionHID) ReadFeature(data []byte) (int, error) {
	h.reads++
	h.readInput = append([]byte(nil), data...)
	copy(data, h.response)
	return h.readSize, h.readErr
}

type transactionQuery struct {
	name     string
	minimum  int
	response []byte
	want     interface{}
	call     func(*Device) (interface{}, error)
}

func transactionQueries() []transactionQuery {
	return []transactionQuery{
		{
			name: "RGB", minimum: 5,
			response: []byte{0, 'r', 12, 34, 56, 0, 0, 0, 0}, want: [3]byte{12, 34, 56},
			call: func(d *Device) (interface{}, error) {
				r, g, b, err := d.ReadRGB(LED2)
				return [3]byte{r, g, b}, err
			},
		},
		{
			name: "play state", minimum: 7,
			response: []byte{0, 'S', 1, 2, 5, 3, 4, 0, 0},
			want:     DevicePatternState{IsPlaying: true, LoopStartPos: 2, LoopEndPos: 5, RepeatTimes: 3, CurrentPos: 4},
			call:     func(d *Device) (interface{}, error) { return d.ReadPlaystate() },
		},
		{
			name: "pattern line", minimum: 8,
			response: []byte{0, 'R', 12, 34, 56, 0, 2, 2, 0},
			want:     DeviceLightState{R: 12, G: 34, B: 56, FadeTimeMsec: 20, LED: LED2},
			call:     func(d *Device) (interface{}, error) { return d.ReadPatternLine(12) },
		},
		{
			name: "version", minimum: 5,
			response: []byte{0, 'v', 0, '2', '4', 0, 0, 0, 0}, want: 204,
			call: func(d *Device) (interface{}, error) { return d.GetVersion() },
		},
		{
			// Test response markers differ by firmware and do not echo '!'.
			name: "raw test", minimum: 1,
			response: []byte{0x55, 0xAA, 1, 2, 3, 4, 5, 6, 0},
			call:     func(d *Device) (interface{}, error) { return d.Test() },
		},
	}
}

func TestDeviceQueryWriteFailure(t *testing.T) {
	writeErr := errors.New("request was not sent")
	for _, query := range transactionQueries() {
		t.Run(query.name, func(t *testing.T) {
			h := &transactionHID{writeErr: writeErr, response: query.response, readSize: 8}
			d := &Device{gen: 2, dev: h}
			if _, err := query.call(d); !errors.Is(err, writeErr) {
				t.Fatalf("query error = %v, want write failure", err)
			}
			if len(h.writes) != 1 || h.reads != 0 {
				t.Fatalf("failed request performed %d writes and %d reads", len(h.writes), h.reads)
			}
		})
	}
}

func TestDeviceQueryReadFailure(t *testing.T) {
	readErr := errors.New("native read failed")
	for _, query := range transactionQueries() {
		t.Run(query.name, func(t *testing.T) {
			h := &transactionHID{readErr: readErr, response: query.response, readSize: 0}
			d := &Device{gen: 2, dev: h}
			if _, err := query.call(d); !errors.Is(err, readErr) {
				t.Fatalf("query error = %v, want native read failure", err)
			}
			if len(h.writes) != 1 || h.reads != 1 {
				t.Fatalf("query performed %d writes and %d reads", len(h.writes), h.reads)
			}
		})
	}
}

func TestDeviceQueryShortResponse(t *testing.T) {
	for _, query := range transactionQueries() {
		t.Run(query.name, func(t *testing.T) {
			// Filling the whole fake buffer must not conceal a short native n.
			h := &transactionHID{response: query.response, readSize: query.minimum - 1}
			d := &Device{gen: 2, dev: h}
			if _, err := query.call(d); !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("short query error = %v, want unexpected EOF", err)
			}
		})
	}
}

func TestDeviceQueryRejectsInvalidNativeCount(t *testing.T) {
	for _, count := range []int{-1, cmdBufSize + 1} {
		h := &transactionHID{readSize: count}
		d := &Device{dev: h}
		if _, err := d.GetVersion(); err == nil {
			t.Fatalf("native count %d was accepted", count)
		}
	}
}

func TestDeviceQueryResponseLengths(t *testing.T) {
	for _, query := range transactionQueries() {
		t.Run(query.name, func(t *testing.T) {
			for _, size := range []int{query.minimum, 8, 9} {
				h := &transactionHID{response: query.response, readSize: size}
				d := &Device{gen: 2, dev: h}
				got, err := query.call(d)
				if err != nil {
					t.Fatalf("response length %d: %v", size, err)
				}
				if query.want != nil && !reflect.DeepEqual(got, query.want) {
					t.Fatalf("response length %d: got %#v, want %#v", size, got, query.want)
				}
				if query.want == nil {
					// Keep the public raw Test buffer size while zeroing bytes
					// beyond the actual response length.
					want := make([]byte, cmdBufSize)
					copy(want, query.response[:size])
					if !bytes.Equal(got.([]byte), want) {
						t.Fatalf("raw response length %d: got %x, want %x", size, got, want)
					}
				}
			}
		})
	}
}

func TestDeviceQueryDoesNotReuseRequestAsResponse(t *testing.T) {
	h := &transactionHID{readSize: 8}
	d := &Device{gen: 2, dev: h}
	st, err := d.ReadPatternLine(12)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.writes) != 1 || h.writes[0][1] != 'R' || h.writes[0][7] != 12 {
		t.Fatalf("wrong pattern request: %x", h.writes)
	}
	wantRead := make([]byte, cmdBufSize)
	wantRead[0] = reportID
	if !bytes.Equal(h.readInput, wantRead) || st.LED != LEDAll {
		t.Fatalf("request data leaked into response: input=%x, state=%#v", h.readInput, st)
	}
}
