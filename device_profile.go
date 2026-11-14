package blink1

import (
	"strings"

	hid "github.com/b1ug/gid"
)

const (
	ProfileModelUnknown = "unknown"
	ProfileModelMK2     = "mk2"
	ProfileModelMK3     = "mk3"
)

// ProfileCapability describes a report format from the referenced protocol.
// It does not enable an SDK operation or establish support by a HID backend.
type ProfileCapability struct {
	Name             string `json:"name"`
	ReportID         byte   `json:"report_id"`
	HostBufferLength int    `json:"host_buffer_length"`
	Evidence         string `json:"evidence"`
	HardwareVerified bool   `json:"hardware_verified"`
	Source           string `json:"source"`
}

// DeviceProfile is a metadata-only identification result. USBDeviceVersion is
// the raw bcdDevice field, not a generation number. Capabilities are protocol
// hints; hardware verification must be recorded independently for each actual
// device, firmware, host platform, backend and operation.
type DeviceProfile struct {
	Model            string              `json:"model"`
	Generation       uint16              `json:"generation"`
	USBDeviceVersion uint16              `json:"usb_device_version"`
	SerialNumber     string              `json:"serial"`
	FirmwareVersion  int                 `json:"firmware,omitempty"`
	IdentityEvidence string              `json:"identity_evidence"`
	Reason           string              `json:"reason"`
	Capabilities     []ProfileCapability `json:"capabilities"`
}

// IdentifyDeviceProfile inspects supplied metadata without enumerating or
// opening hardware. Pass firmwareVersion=0 when it has not been queried. Known
// model identification requires consistent VID/PID, serial prefix, product and
// bcdDevice; supplied firmware must agree too. Missing, unfamiliar or conflicting
// data yields an unknown profile with no capability hints.
//
// This additive helper does not change OpenDevice or GetGeneration. In
// particular, the mk3 descriptor's bcdDevice=0x0101 is retained as raw USB
// metadata while a consistently identified mk3 has profile generation 3.
func IdentifyDeviceProfile(info *hid.DeviceInfo, firmwareVersion int) DeviceProfile {
	profile := DeviceProfile{
		Model: ProfileModelUnknown, FirmwareVersion: firmwareVersion,
		IdentityEvidence: "metadata-only", Reason: "missing-metadata",
		Capabilities: []ProfileCapability{},
	}
	if info == nil {
		return profile
	}
	profile.USBDeviceVersion = info.VersionNumber
	profile.SerialNumber = info.SerialNumber
	if !IsBlink1Device(info) {
		profile.Reason = "not-blink1-vid-pid"
		return profile
	}
	if !validProfileSerial(info.SerialNumber) {
		profile.Reason = "missing-or-invalid-serial"
		return profile
	}
	var model, product, source string
	var generation, usbVersion uint16
	switch info.SerialNumber[0] {
	case '2':
		model, product, generation, usbVersion = ProfileModelMK2, "blink(1) mk2", 2, 0x0002
		source = "https://github.com/todbot/blink1/blob/101abe3599c661c4dd4d1204b4d50fb592293143/hardware/firmware_mk2/usb_descriptors.c"
	case '3':
		model, product, generation, usbVersion = ProfileModelMK3, "blink(1) mk3", 3, 0x0101
		source = "https://github.com/todbot/blink1mk3/blob/ccf04e26726a4a41ef9689653ae12baa73dbc456/firmware/firmware-v30x/descriptors.h"
	default:
		profile.Reason = "unknown-serial-prefix"
		return profile
	}
	if !strings.EqualFold(strings.TrimSpace(info.Product), product) {
		profile.Reason = "missing-or-conflicting-product"
		return profile
	}
	if info.VersionNumber != usbVersion {
		profile.Reason = "missing-or-conflicting-usb-version"
		return profile
	}
	if firmwareVersion != 0 {
		// GetVersion has two single ASCII digits, encoded as major*100+minor.
		// An integer such as 217 cannot come from that validated format.
		if firmwareVersion < 0 || firmwareVersion%100 > 9 || firmwareVersion/100 != int(generation) {
			profile.Reason = "invalid-or-conflicting-firmware"
			return profile
		}
	}
	profile.Model, profile.Generation, profile.Reason = model, generation, "consistent-known-metadata"
	profile.Capabilities = append(profile.Capabilities, ProfileCapability{
		Name: "basic-feature-report", ReportID: reportID, HostBufferLength: cmdBufSize,
		Evidence: "protocol-inferred", HardwareVerified: false, Source: source,
	})
	if model == ProfileModelMK3 {
		profile.Capabilities = append(profile.Capabilities, ProfileCapability{
			Name: "extended-feature-report", ReportID: report3ID, HostBufferLength: cmdBuf3Size,
			Evidence: "protocol-inferred", HardwareVerified: false, Source: source,
		})
	}
	return profile
}

func validProfileSerial(serial string) bool {
	if len(serial) != 8 {
		return false
	}
	for _, c := range serial {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
