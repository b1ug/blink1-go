package blink1

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	hid "github.com/b1ug/gid"
)

type inspectionHID struct {
	transactionHID
	returnedBuffer []byte
	readStarted    chan struct{}
	releaseRead    chan struct{}
	closed         chan struct{}
}

func (h *inspectionHID) ReadFeature(data []byte) (int, error) {
	if h.readStarted != nil {
		close(h.readStarted)
		<-h.releaseRead
	}
	h.returnedBuffer = data
	return h.transactionHID.ReadFeature(data)
}
func (h *inspectionHID) Close() {
	if h.closed != nil {
		close(h.closed)
	}
}

func inspectionDevice(h *inspectionHID, model string) *Device {
	info := patternTestMetadata(model)
	return &Device{gen: info.VersionNumber, pn: info.Product, sn: info.SerialNumber, info: info, dev: h}
}

func inspectionQueries() []struct {
	name string
	min  int
	call func(*Device) (RawFeatureResponse, error)
} {
	return []struct {
		name string
		min  int
		call func(*Device) (RawFeatureResponse, error)
	}{
		{"R", 8, func(d *Device) (RawFeatureResponse, error) { r, e := d.ReadPatternLineRaw(12); return r.Response, e }},
		{"S", 7, func(d *Device) (RawFeatureResponse, error) { r, e := d.ReadPlaystateRaw(); return r.Response, e }},
	}
}

func TestRawPatternLinePreservesZeroUnknownLEDAndRawTime(t *testing.T) {
	for _, model := range []string{ProfileModelMK2, ProfileModelMK3} {
		for _, led := range []byte{0, 1, 2, 0x80, 0xff} {
			t.Run(fmt.Sprintf("%s/led%d", model, led), func(t *testing.T) {
				data := []byte{0x42, 0x19, 12, 34, 56, 0xff, 0xff, led, 0x99}
				h := &inspectionHID{transactionHID: transactionHID{response: data, readSize: len(data)}}
				got, err := inspectionDevice(h, model).ReadPatternLineRaw(31)
				if err != nil || got.Position != 31 || got.RGB8 == nil || *got.RGB8 != [3]byte{12, 34, 56} || got.DurationTicks == nil || *got.DurationTicks != 65535 || got.LEDCode == nil || *got.LEDCode != led {
					t.Fatalf("raw line changed: %#v, %v", got, err)
				}
				if !bytes.Equal(got.Response.Bytes(), data) || !bytes.Equal(got.Response.UnknownTail(), []byte{0x99}) {
					t.Fatal("response markers or tail changed")
				}
				if len(h.writes) != 1 || !bytes.Equal(h.writes[0], []byte{1, 'R', 0, 0, 0, 0, 0, 31, 0}) || h.reads != 1 || !bytes.Equal(h.readInput, []byte{1, 0, 0, 0, 0, 0, 0, 0, 0}) {
					t.Fatal("raw read retried, leaked request data, or sent another command", h.writes, h.readInput)
				}
			})
		}
	}
	for _, position := range []uint{0, 15, 16, 31} {
		h := &inspectionHID{transactionHID: transactionHID{readSize: 8}}
		got, err := inspectionDevice(h, ProfileModelMK2).ReadPatternLineRaw(position)
		if err != nil || got.Position != position || *got.RGB8 != [3]byte{} || *got.DurationTicks != 0 || *got.LEDCode != 0 || h.writes[0][7] != byte(position) {
			t.Fatal("zero RAM row was rejected, omitted or normalized", position, got, err)
		}
	}
}

func TestRawPlaystatePreservesModesCountsAndCursor(t *testing.T) {
	for _, model := range []string{ProfileModelMK2, ProfileModelMK3} {
		for _, mode := range []byte{0, 1, 2, 3, 0x7f, 0xff} {
			for _, count := range []byte{0, 1, 255} {
				t.Run(fmt.Sprintf("%s/mode%d/count%d", model, mode, count), func(t *testing.T) {
					data := []byte{0, 'S', mode, 31, 32, count, 0, 0xa5, 0x5a}
					h := &inspectionHID{transactionHID: transactionHID{response: data, readSize: len(data)}}
					got, err := inspectionDevice(h, model).ReadPlaystateRaw()
					if err != nil || *got.Mode != mode || *got.Start != 31 || *got.End != 32 || *got.Count != count || *got.NextPosition != 0 {
						t.Fatalf("raw state changed: %#v, %v", got, err)
					}
					if !bytes.Equal(got.Response.Bytes(), data) || !bytes.Equal(got.Response.UnknownTail(), []byte{0xa5, 0x5a}) || len(h.writes) != 1 || !bytes.Equal(h.writes[0], []byte{1, 'S', 0, 0, 0, 0, 0, 0, 0}) {
						t.Fatal("raw state wire/tail changed")
					}
				})
			}
		}
	}
}

func TestRawInspectionRejectsUnknownProfileBeforeIO(t *testing.T) {
	for _, query := range inspectionQueries() {
		for _, kind := range []string{"nil", "generation-only", "mk1", "conflicting-product", "conflicting-usb-version", "conflicting-vid"} {
			t.Run(query.name+"/"+kind, func(t *testing.T) {
				h := &inspectionHID{}
				d := inspectionDevice(h, ProfileModelMK2)
				switch kind {
				case "nil":
					d.info = nil
				case "generation-only":
					d.info = &hid.DeviceInfo{VersionNumber: 2}
				case "mk1":
					d.info.SerialNumber = "10001234"
				case "conflicting-product":
					d.info.Product = "unknown"
				case "conflicting-usb-version":
					d.info.VersionNumber = 0x9999
				case "conflicting-vid":
					d.info.VendorID = 0
				}
				r, err := query.call(d)
				if !errors.Is(err, ErrRawInspectionUnsupported) || len(h.writes) != 0 || h.reads != 0 || r.ReadAttempted || len(r.Bytes()) != 0 || r.ExpectedFieldsAvailable() {
					t.Fatal("unknown profile probed or became a result", r, err)
				}
			})
		}
	}
}

func TestRawInspectionRejectsRangeBeforeIO(t *testing.T) {
	for _, position := range []uint{32, 33, 255, 256, ^uint(0)} {
		t.Run(fmt.Sprint(position), func(t *testing.T) {
			h := &inspectionHID{}
			got, err := inspectionDevice(h, ProfileModelMK2).ReadPatternLineRaw(position)
			if !errors.Is(err, ErrRawInspectionPosition) || got.Position != position || got.Response.ReadAttempted || len(h.writes) != 0 || h.reads != 0 {
				t.Fatal("out-of-range position aliased or performed I/O", got, err)
			}
		})
	}
}

func TestRawInspectionEveryResponsePrefix(t *testing.T) {
	for _, query := range inspectionQueries() {
		for n := 0; n <= cmdBufSize; n++ {
			t.Run(fmt.Sprintf("%s/n%d", query.name, n), func(t *testing.T) {
				data := []byte{9, 8, 7, 6, 5, 4, 3, 2, 1}
				h := &inspectionHID{transactionHID: transactionHID{response: data, readSize: n}}
				r, err := query.call(inspectionDevice(h, ProfileModelMK2))
				if (err == nil) != (n >= query.min) || (n < query.min && !errors.Is(err, io.ErrUnexpectedEOF)) {
					t.Fatal("wrong short-read handling", n, err)
				}
				if !r.ReadAttempted || !r.LengthValid || r.ReturnedLength != n || r.BufferLength != cmdBufSize || r.RequestedReportID != reportID || !bytes.Equal(r.Bytes(), data[:n]) || r.ExpectedFieldsAvailable() != (n >= query.min) {
					t.Fatal("invented or dropped prefix", r, r.Bytes())
				}
				var tail []byte
				if n > query.min {
					tail = data[query.min:n]
				}
				if !bytes.Equal(r.UnknownTail(), tail) || len(h.writes) != 1 || h.reads != 1 {
					t.Fatal("tail or retry changed")
				}
			})
		}
	}
}

func TestRawInspectionPartialFieldAvailability(t *testing.T) {
	for n := 0; n <= cmdBufSize; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			h := &inspectionHID{transactionHID: transactionHID{response: []byte{0, 'R', 1, 2, 3, 0, 5, 255, 6}, readSize: n}}
			got, _ := inspectionDevice(h, ProfileModelMK2).ReadPatternLineRaw(3)
			if (got.RGB8 != nil) != (n >= 5) || (got.DurationTicks != nil) != (n >= 7) || (got.LEDCode != nil) != (n >= 8) {
				t.Fatal("partial group fabricated a value", got)
			}
			h = &inspectionHID{transactionHID: transactionHID{response: []byte{0, 'S', 0, 0, 0, 0, 0, 0, 0}, readSize: n}}
			st, _ := inspectionDevice(h, ProfileModelMK2).ReadPlaystateRaw()
			for i, field := range []*byte{st.Mode, st.Start, st.End, st.Count, st.NextPosition} {
				if (field != nil) != (n > i+2) {
					t.Fatal("partial state fabricated a value", n, i)
				}
			}
		})
	}
}

func TestRawInspectionInvalidCountsAndNativeErrorChains(t *testing.T) {
	cause := errors.New("native cause")
	native := &hid.NativeError{Platform: "darwin", Code: 0xe00002bc, Err: cause}
	for _, query := range inspectionQueries() {
		for _, n := range []int{-2, -1, 0, 1, 4, 6, 8, 9, 10, 100} {
			for _, fail := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/n%d/error%v", query.name, n, fail), func(t *testing.T) {
					data := []byte{0, 'R', 1, 2, 3, 4, 5, 6, 7}
					h := &inspectionHID{transactionHID: transactionHID{response: data, readSize: n}}
					if fail {
						h.readErr = native
					}
					r, err := query.call(inspectionDevice(h, ProfileModelMK2))
					valid := n >= 0 && n <= cmdBufSize
					if r.LengthValid != valid || r.ReturnedLength != n || !r.ReadAttempted || len(h.writes) != 1 || h.reads != 1 {
						t.Fatal("length or retry changed", r)
					}
					if valid {
						if !bytes.Equal(r.Bytes(), data[:n]) {
							t.Fatal("partial native response dropped or extended", r.Bytes())
						}
					} else if len(r.Bytes()) != 0 || r.ExpectedFieldsAvailable() || err == nil {
						t.Fatal("invalid n became trusted bytes", r, err)
					}
					if fail {
						var found *hid.NativeError
						var carrier interface {
							NativePlatform() string
							NativeCode() uint64
						}
						if !errors.Is(err, cause) || !errors.Is(err, native) || !errors.As(err, &found) || found != native || !errors.As(err, &carrier) || carrier.NativeCode() != native.Code {
							t.Fatal("native cause/type/carrier lost", err)
						}
					}
				})
			}
		}
	}
}

func TestRawInspectionWriteAndClosedErrorsDoNotRead(t *testing.T) {
	for _, query := range inspectionQueries() {
		t.Run(query.name, func(t *testing.T) {
			cause := errors.New("write request failed")
			native := &hid.NativeError{Platform: "darwin", Code: 17, Err: cause}
			h := &inspectionHID{transactionHID: transactionHID{writeErr: native}}
			d := inspectionDevice(h, ProfileModelMK2)
			r, err := query.call(d)
			var typed *hid.NativeError
			if !errors.Is(err, cause) || !errors.As(err, &typed) || typed != native || r.ReadAttempted || len(r.Bytes()) != 0 || len(h.writes) != 1 || h.reads != 0 {
				t.Fatal("failed request read or lost cause", r, err)
			}
			d.Close()
			r, err = query.call(d)
			if !errors.Is(err, errDeviceClosed) || r.ReadAttempted || len(h.writes) != 1 || h.reads != 0 {
				t.Fatal("closed handle performed I/O", r, err)
			}
		})
	}
}

func TestRawInspectionOwnsReturnedFieldsAndByteViews(t *testing.T) {
	data := []byte{0, 'R', 10, 20, 30, 0, 1, 255, 0x91}
	h := &inspectionHID{transactionHID: transactionHID{response: data, readSize: 9}}
	d := inspectionDevice(h, ProfileModelMK2)
	first, err := d.ReadPatternLineRaw(1)
	if err != nil {
		t.Fatal(err)
	}
	for i := range h.returnedBuffer {
		h.returnedBuffer[i] = 0
	}
	view := first.Response.Bytes()
	view[2] = 99
	tail := first.Response.UnknownTail()
	tail[0] = 0
	if !bytes.Equal(first.Response.Bytes(), data) || !bytes.Equal(first.Response.UnknownTail(), []byte{0x91}) || *first.RGB8 != [3]byte{10, 20, 30} {
		t.Fatal("backend or accessor shares data")
	}
	second, err := d.ReadPatternLineRaw(1)
	if err != nil {
		t.Fatal(err)
	}
	first.RGB8[0], *first.DurationTicks, *first.LEDCode = 99, 77, 42
	if *second.RGB8 != [3]byte{10, 20, 30} || *second.DurationTicks != 1 || *second.LEDCode != 255 {
		t.Fatal("line results share parsed pointers")
	}
	h.response = []byte{0, 'S', 3, 4, 5, 6, 7, 8, 9}
	a, _ := d.ReadPlaystateRaw()
	b, _ := d.ReadPlaystateRaw()
	*a.Mode, *a.Start, *a.End, *a.Count, *a.NextPosition = 0, 0, 0, 0, 0
	if *b.Mode != 3 || *b.Start != 4 || *b.End != 5 || *b.Count != 6 || *b.NextPosition != 7 {
		t.Fatal("state results share parsed pointers")
	}
}

func TestRawInspectionHoldsDeviceLeaseUntilResponseReturns(t *testing.T) {
	h := &inspectionHID{transactionHID: transactionHID{response: make([]byte, 9), readSize: 9}, readStarted: make(chan struct{}), releaseRead: make(chan struct{}), closed: make(chan struct{})}
	d := inspectionDevice(h, ProfileModelMK2)
	done := make(chan error, 1)
	go func() { _, err := d.ReadPatternLineRaw(1); done <- err }()
	released := false
	defer func() {
		if !released {
			close(h.releaseRead)
		}
	}()
	select {
	case <-h.readStarted:
	case <-time.After(time.Second):
		t.Fatal("read did not reach fake transport")
	}
	closeDone := make(chan struct{})
	go func() { d.Close(); close(closeDone) }()
	select {
	case <-h.closed:
		t.Fatal("Close released handle during response")
	case <-time.After(20 * time.Millisecond):
	}
	close(h.releaseRead)
	released = true
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("read did not finish")
	}
	select {
	case <-closeDone:
	case <-time.After(time.Second):
		t.Fatal("Close did not complete after response")
	}
	if _, err := d.ReadPlaystateRaw(); !errors.Is(err, errDeviceClosed) {
		t.Fatal("closed device became readable", err)
	}
}
