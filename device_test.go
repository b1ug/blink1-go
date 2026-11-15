package blink1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	hid "github.com/b1ug/gid"
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

func TestDeviceFailedRequestDoesNotWaitForResponse(t *testing.T) {
	writeErr := errors.New("request was not sent")
	h := &transactionHID{writeErr: writeErr}
	d := &Device{dev: h}
	done := make(chan error, 1)
	go func() {
		done <- d.delayRead([]byte{1, '!'}, 60*1000, 1)
	}()
	select {
	case err := <-done:
		if !errors.Is(err, writeErr) || h.reads != 0 {
			t.Fatalf("failed request = %v, reads = %d", err, h.reads)
		}
	case <-time.After(time.Second):
		t.Fatal("failed request waited for the device response delay")
	}
}

// firmwareVersionHID replies using the same report layout as the firmware's
// 'v' handler, without opening a USB device.
type firmwareVersionHID struct {
	major byte
	minor byte
}

func (h *firmwareVersionHID) Close()                         {}
func (h *firmwareVersionHID) Write(data []byte) error        { return nil }
func (h *firmwareVersionHID) WriteFeature(data []byte) error { return nil }
func (h *firmwareVersionHID) Read(data []byte) (int, error) {
	return h.ReadFeature(data)
}
func (h *firmwareVersionHID) ReadFeature(data []byte) (int, error) {
	return copy(data, []byte{1, 'v', 0, h.major, h.minor, 0, 0, 0, 0}), nil
}

func TestGetVersion(t *testing.T) {
	for _, tc := range []struct {
		name         string
		major, minor byte
		want         int
	}{
		{name: "mk2 204", major: '2', minor: '4', want: 204},
		{name: "mk3 303", major: '3', minor: '3', want: 303},
		{name: "mk3 304", major: '3', minor: '4', want: 304},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &Device{dev: &firmwareVersionHID{major: tc.major, minor: tc.minor}}
			got, err := d.GetVersion()
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("GetVersion() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestGetVersionRejectsNonDigits(t *testing.T) {
	for _, tc := range []struct {
		name         string
		major, minor byte
	}{
		{"major before zero", '/', '4'},
		{"major after nine", ':', '4'},
		{"major NUL", 0, '4'},
		{"minor after nine looks like 210", '2', ':'},
		{"minor letter looks like 217", '2', 'A'},
		{"minor NUL", '2', 0},
		{"minor non ASCII", '2', 0xff},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &Device{dev: &firmwareVersionHID{major: tc.major, minor: tc.minor}}
			got, err := d.GetVersion()
			if err == nil || got != 0 {
				t.Fatalf("malformed digits produced version %d, error %v", got, err)
			}
		})
	}
}

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

func TestDeviceProfileGoldenFixtures(t *testing.T) {
	var fixtures []struct {
		Name     string          `json:"name"`
		Metadata *hid.DeviceInfo `json:"metadata"`
		Firmware int             `json:"firmware"`
		Want     struct {
			Model         string `json:"model"`
			Generation    uint16 `json:"generation"`
			USBVersion    uint16 `json:"usb_version"`
			Reason        string `json:"reason"`
			ReportIDs     []byte `json:"report_ids"`
			BufferLengths []int  `json:"buffer_lengths"`
		} `json:"want"`
	}
	data, err := ioutil.ReadFile("testdata/device_profiles.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			got := IdentifyDeviceProfile(fixture.Metadata, fixture.Firmware)
			if got.Model != fixture.Want.Model || got.Generation != fixture.Want.Generation || got.USBDeviceVersion != fixture.Want.USBVersion || got.Reason != fixture.Want.Reason {
				t.Fatalf("profile = %#v, want %#v", got, fixture.Want)
			}
			if got.SerialNumber != fixture.Metadata.SerialNumber || got.FirmwareVersion != fixture.Firmware || got.IdentityEvidence != "metadata-only" {
				t.Fatalf("profile lost metadata or promoted evidence: %#v", got)
			}
			ids, lengths := []byte{}, []int{}
			for _, capability := range got.Capabilities {
				ids, lengths = append(ids, capability.ReportID), append(lengths, capability.HostBufferLength)
				if capability.Evidence != "protocol-inferred" || capability.HardwareVerified || capability.Source == "" {
					t.Fatalf("metadata created a verification claim: %#v", capability)
				}
			}
			if !reflect.DeepEqual(ids, fixture.Want.ReportIDs) || !reflect.DeepEqual(lengths, fixture.Want.BufferLengths) {
				t.Fatalf("reports = %v/%v, want %v/%v", ids, lengths, fixture.Want.ReportIDs, fixture.Want.BufferLengths)
			}
		})
	}
}

func TestDeviceProfileIncompleteMetadata(t *testing.T) {
	for _, info := range []*hid.DeviceInfo{
		nil,
		{VendorID: b1VendorID, ProductID: b1ProductID, VersionNumber: 2, Product: "blink(1) mk2"},
		{VendorID: b1VendorID, ProductID: b1ProductID, VersionNumber: 2, Product: "blink(1) mk2", SerialNumber: "2000560"},
		{VendorID: b1VendorID, ProductID: b1ProductID, VersionNumber: 2, Product: "blink(1) mk2", SerialNumber: "2ggggggg"},
	} {
		profile := IdentifyDeviceProfile(info, 204)
		if profile.Model != ProfileModelUnknown || profile.Generation != 0 || len(profile.Capabilities) != 0 {
			t.Fatalf("incomplete metadata produced a known profile: %#v", profile)
		}
	}
}

func TestDeviceProfileDoesNotChangeLegacyGeneration(t *testing.T) {
	info := &hid.DeviceInfo{VendorID: b1VendorID, ProductID: b1ProductID, VersionNumber: 0x0101, Product: "blink(1) mk3", SerialNumber: "30001234"}
	d := &Device{info: info, gen: info.VersionNumber}
	profile := IdentifyDeviceProfile(info, 304)
	if profile.Generation != 3 || profile.USBDeviceVersion != 0x0101 || d.GetGeneration() != 0x0101 || info.VersionNumber != 0x0101 {
		t.Fatalf("profile changed legacy state: profile=%#v, generation=%d, metadata=%#v", profile, d.GetGeneration(), info)
	}
	profile.Capabilities[0].HardwareVerified = true
	if next := IdentifyDeviceProfile(info, 304); next.Capabilities[0].HardwareVerified {
		t.Fatal("profiles share mutable capability state")
	}
}
