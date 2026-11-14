package blink1

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

// savePatternHID records the save report without enumerating or opening USB.
type savePatternHID struct {
	writeErr error
	writes   [][]byte
	reads    int
}

func (h *savePatternHID) Close()                  {}
func (h *savePatternHID) Write(data []byte) error { return h.WriteFeature(data) }
func (h *savePatternHID) WriteFeature(data []byte) error {
	h.writes = append(h.writes, append([]byte(nil), data...))
	return h.writeErr
}
func (h *savePatternHID) Read(data []byte) (int, error) { return h.ReadFeature(data) }
func (h *savePatternHID) ReadFeature(data []byte) (int, error) {
	h.reads++
	return len(data), nil
}

type savePatternTransportError struct {
	code  int
	cause error
}

func (e *savePatternTransportError) Error() string { return e.cause.Error() }
func (e *savePatternTransportError) Unwrap() error { return e.cause }

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
	typedErr := &savePatternTransportError{code: 123, cause: transportErr}
	for _, caller := range savePatternCallers() {
		t.Run(caller.name, func(t *testing.T) {
			for _, test := range []struct {
				name string
				err  error
			}{
				{"success", nil},
				{"transport_error", transportErr},
				{"timeout", context.DeadlineExceeded},
				{"typed_transport_error", typedErr},
			} {
				t.Run(test.name, func(t *testing.T) {
					h := &savePatternHID{writeErr: test.err}
					d := &Device{gen: 2, dev: h}
					err := caller.call(d)
					if !errors.Is(err, test.err) {
						t.Errorf("save error = %v, want cause %v", err, test.err)
					}
					if test.err == typedErr {
						var got *savePatternTransportError
						if !errors.As(err, &got) || got != typedErr || got.code != 123 || !errors.Is(err, transportErr) {
							t.Errorf("save lost typed transport error identity: %v", err)
						}
					}
					want := []byte{1, 'W', 0xBE, 0xEF, 0xCA, 0xFE, 0, 0, 0}
					if len(h.writes) != 1 || !bytes.Equal(h.writes[0], want) || h.reads != 0 {
						t.Errorf("save I/O = writes %x, reads %d; want one original report", h.writes, h.reads)
					}
				})
			}
		})
	}
}

// appliedSaveTimeoutHID models a save taking effect before the transport reports
// a timeout. It is not a flash or power-loss durability simulation.
type appliedSaveTimeoutHID struct {
	savePatternHID
	saveApplied bool
}

func (h *appliedSaveTimeoutHID) WriteFeature(data []byte) error {
	if bytes.Equal(data, []byte{1, 'W', 0xBE, 0xEF, 0xCA, 0xFE, 0, 0, 0}) {
		h.saveApplied = true
	}
	return h.savePatternHID.WriteFeature(data)
}

func TestSavePatternAppliedBeforeTimeout(t *testing.T) {
	for _, caller := range savePatternCallers() {
		t.Run(caller.name, func(t *testing.T) {
			h := &appliedSaveTimeoutHID{savePatternHID: savePatternHID{writeErr: context.DeadlineExceeded}}
			d := &Device{gen: 2, dev: h}
			err := caller.call(d)
			if !h.saveApplied || !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("applied = %v, error = %v; want the communication error even after the modeled save", h.saveApplied, err)
			}
			if len(h.writes) != 1 || h.reads != 0 {
				t.Errorf("ambiguous save was repeated or queried: writes %x, reads %d", h.writes, h.reads)
			}
		})
	}
}
