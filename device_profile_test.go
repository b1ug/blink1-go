package blink1

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	hid "github.com/b1ug/gid"
)

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
