package blink1

import (
	"errors"
	"fmt"
	"io"
)

var (
	// ErrRawInspectionUnsupported means consistent known MK2/MK3 metadata and
	// a known RAM capacity were unavailable. No feature request was sent.
	ErrRawInspectionUnsupported = errors.New("b1: raw Pattern inspection requires a known MK2 or MK3 profile")
	// ErrRawInspectionPosition means the requested RAM position was outside
	// the known capacity. Firmware must not silently alias it to position zero.
	ErrRawInspectionPosition = errors.New("b1: raw Pattern position is outside the known RAM capacity")
)

// RawFeatureResponse records exactly the prefix reported by GID ReadFeature,
// not a complete native USB wire capture. BufferLength includes GID's ID slot;
// RequestedReportID selects that report. A returned slot is not independently
// authenticated as a native report ID or command echo.
//
// ReturnedLength is meaningful only when ReadAttempted is true. LengthValid
// means 0 <= ReturnedLength <= BufferLength. Invalid counts retain no trusted
// bytes. A valid count, including a partial count returned with an error, retains
// only that prefix; bytes outside it are not retained, and the prefilled ID does
// not create returned data when n=0. A backend that returns zero on error exposes
// no trusted bytes. Within n, this API cannot distinguish a backend-retained ID
// slot from one overwritten by the native read.
// All returned buffers and parsed fields belong to this result alone.
type RawFeatureResponse struct {
	RequestedReportID byte
	BufferLength      int
	ReadAttempted     bool
	ReturnedLength    int
	LengthValid       bool
	data              []byte
	fieldPrefixLength int
}

// Bytes returns a defensive copy of the available GID response prefix.
func (r RawFeatureResponse) Bytes() []byte {
	return append([]byte(nil), r.data...)
}

// UnknownTail returns a defensive copy of bytes after the fields understood by
// this command. Reserved bytes are retained without assigning them semantics.
func (r RawFeatureResponse) UnknownTail() []byte {
	if len(r.data) <= r.fieldPrefixLength {
		return nil
	}
	return append([]byte(nil), r.data[r.fieldPrefixLength:]...)
}

// ExpectedFieldsAvailable reports byte availability, not transport success,
// playback completion, snapshot consistency, or hardware qualification. The
// caller must still check the error returned by the read operation.
func (r RawFeatureResponse) ExpectedFieldsAvailable() bool {
	return r.ReadAttempted && r.LengthValid && r.fieldPrefixLength > 0 && len(r.data) >= r.fieldPrefixLength
}

// RawPatternLine is a single RAM-row observation. Position is the requested
// absolute position; the R response does not echo it. Nil fields were not fully
// returned. RGB8 is raw device RGB, DurationTicks is the unscaled uint16 field,
// and LEDCode preserves slot 7 without LED normalization or gamma. That slot's
// stored-LED meaning is defined by the referenced firmware204+ layout; metadata
// identification alone does not qualify its meaning on a different firmware.
// This contains no host Pattern title, recipe, revision, or Flash verification.
type RawPatternLine struct {
	Position      uint
	RGB8          *[3]byte
	DurationTicks *uint16
	LEDCode       *byte
	Response      RawFeatureResponse
}

// RawPlaystate preserves the S fields without interpreting mode as a bool,
// guessing the current row, or declaring forever/completed. NextPosition is the
// raw cursor; End is the raw firmware value, not a normalized inclusive/exclusive
// boundary. MK2 and MK3 have different end conventions. Nil means unavailable.
type RawPlaystate struct {
	Mode         *byte
	Start        *byte
	End          *byte
	Count        *byte
	NextPosition *byte
	Response     RawFeatureResponse
}

// ReadPatternLineRaw issues one R query, without retries, for a valid position
// in the supplied metadata's known MK2/MK3 RAM capacity. It does not probe unknown
// models, stop playback, upload, save, or enumerate/open a device. Short responses
// preserve their available prefix and return an error. Partial n+error responses
// preserve the valid prefix and wrap the original error for errors.Is/As.
//
// Success is one RAM observation, not Flash readback, an atomic multi-row image,
// a recovered high-level recipe, or operation-specific hardware qualification.
func (b1 *Device) ReadPatternLineRaw(position uint) (RawPatternLine, error) {
	response, err := b1.readPatternRaw('R', position, 8)
	result := RawPatternLine{Position: position, Response: response}
	data := response.data
	if len(data) >= 5 {
		rgb := [3]byte{data[2], data[3], data[4]}
		result.RGB8 = &rgb
	}
	if len(data) >= 7 {
		ticks := uint16(data[5])<<8 | uint16(data[6])
		result.DurationTicks = &ticks
	}
	if len(data) >= 8 {
		led := data[7]
		result.LEDCode = &led
	}
	return result, err
}

// ReadPlaystateRaw issues one S query without retries or normalization. The
// known-profile guard is metadata-only and does not query firmware. A stopped
// sequence scheduler can still have a final fade in progress; callers must not
// derive optical completion from these fields. See RawFeatureResponse for the
// GID length, ID-slot, partial-error, and byte-ownership boundary.
func (b1 *Device) ReadPlaystateRaw() (RawPlaystate, error) {
	response, err := b1.readPatternRaw('S', 0, 7)
	result := RawPlaystate{Response: response}
	fields := []**byte{&result.Mode, &result.Start, &result.End, &result.Count, &result.NextPosition}
	for i, field := range fields {
		if len(response.data) > i+2 {
			value := response.data[i+2]
			*field = &value
		}
	}
	return result, err
}

// This additive path intentionally leaves legacy read/delayRead behavior intact.
// Hold the device lease across the request/response and keep its response buffer
// separate from the request so request positions cannot appear as observations.
func (b1 *Device) readPatternRaw(command byte, position uint, prefix int) (RawFeatureResponse, error) {
	result := RawFeatureResponse{RequestedReportID: reportID, BufferLength: cmdBufSize, fieldPrefixLength: prefix}
	b1.mu.Lock()
	defer b1.mu.Unlock()
	if b1.dev == nil {
		return result, errDeviceClosed
	}
	capabilities := IdentifyPatternCapabilities(b1.info, 0)
	capacity := capabilities.Storage.RAMLines
	if capabilities.Identity.Model == ProfileModelUnknown || capacity.Knowledge != PatternKnowledgeKnown || capacity.Value == nil || *capacity.Value <= 0 {
		return result, ErrRawInspectionUnsupported
	}
	if command == 'R' && position >= uint(*capacity.Value) {
		return result, fmt.Errorf("%w: position %d, capacity %d", ErrRawInspectionPosition, position, *capacity.Value)
	}
	request := make([]byte, cmdBufSize)
	request[0], request[1] = reportID, command
	if command == 'R' {
		request[7] = byte(position)
	}
	if err := b1.dev.WriteFeature(request); err != nil {
		return result, fmt.Errorf("b1: raw %c write request fail: %w", command, err)
	}
	response := make([]byte, cmdBufSize)
	response[0] = reportID
	result.ReadAttempted = true
	n, err := b1.dev.ReadFeature(response)
	result.ReturnedLength = n
	result.LengthValid = n >= 0 && n <= len(response)
	if result.LengthValid {
		result.data = append([]byte(nil), response[:n]...)
	} else if err != nil {
		return result, fmt.Errorf("b1: raw %c invalid response length %d for buffer %d; read fail: %w", command, n, len(response), err)
	} else {
		return result, fmt.Errorf("b1: raw %c invalid response length %d for buffer %d", command, n, len(response))
	}
	if err != nil {
		return result, fmt.Errorf("b1: raw %c read fail: %w", command, err)
	}
	if n < prefix {
		return result, fmt.Errorf("b1: short raw %c response: got %d bytes, need at least %d: %w", command, n, prefix, io.ErrUnexpectedEOF)
	}
	return result, nil
}
