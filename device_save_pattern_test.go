package blink1

import (
	"bytes"
	"context"
	"errors"
	"testing"

	hid "github.com/b1ug/gid"
)

func savePatternCallers() []struct {
	name string
	call func(*Device) error
} {
	return []struct {
		name string
		call func(*Device) error
	}{
		{"Device", func(d *Device) error { return d.SavePattern() }},
		{"Controller", func(d *Device) error { return NewController(d).WritePattern() }},
	}
}

func TestSavePatternWriteOutcome(t *testing.T) {
	transportErr := errors.New("fake feature write failed")
	nativeErr := &hid.NativeError{Platform: "test-platform", Code: 123, Err: transportErr}
	for _, caller := range savePatternCallers() {
		t.Run(caller.name, func(t *testing.T) {
			for _, test := range []struct {
				name string
				err  error
			}{
				{"success", nil},
				{"transport_error", transportErr},
				{"timeout", context.DeadlineExceeded},
				{"native_error", nativeErr},
			} {
				t.Run(test.name, func(t *testing.T) {
					h := &transactionHID{writeErr: test.err}
					d := &Device{gen: 2, dev: h}
					err := caller.call(d)
					if !errors.Is(err, test.err) {
						t.Errorf("save error = %v, want cause %v", err, test.err)
					}
					if test.err == nativeErr {
						var got *hid.NativeError
						if !errors.As(err, &got) || got != nativeErr {
							t.Errorf("save lost native error identity: %v", err)
						}
					}
					// Saving may already have reached flash when the transport
					// reports failure. Never retry or add a verification read.
					want := []byte{1, 'W', 0xBE, 0xEF, 0xCA, 0xFE, 0, 0, 0}
					if len(h.writes) != 1 || !bytes.Equal(h.writes[0], want) || h.reads != 0 {
						t.Errorf("save I/O = writes %x, reads %d; want one original report", h.writes, h.reads)
					}
				})
			}
		})
	}
}

func TestSavePatternAfterClose(t *testing.T) {
	for _, caller := range savePatternCallers() {
		t.Run(caller.name, func(t *testing.T) {
			h := &transactionHID{}
			d := &Device{gen: 2, dev: h}
			d.Close()
			if err := caller.call(d); err != errDeviceClosed {
				t.Errorf("closed save error = %v, want %v", err, errDeviceClosed)
			}
			if len(h.writes) != 0 || h.reads != 0 {
				t.Errorf("closed save accessed transport: writes %x, reads %d", h.writes, h.reads)
			}
		})
	}
}
