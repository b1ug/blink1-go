package blink1

import (
	"fmt"

	hid "github.com/b1ug/gid"
)

func ExampleIdentifyPatternCapabilities() {
	// Metadata and firmware are supplied; this example never opens a device.
	info := &hid.DeviceInfo{
		VendorID: 0x27b8, ProductID: 0x01ed,
		VersionNumber: 2, Product: "blink(1) mk2", SerialNumber: "20005601",
	}
	profile := IdentifyPatternCapabilities(info, 204)
	fmt.Println(profile.Identity.FirmwareVersion, *profile.ReferenceFirmwareVersion, profile.Applicability)
	fmt.Println(*profile.Storage.RAMLines.Value, *profile.Storage.FlashLines.Value)
	fmt.Println(profile.Storage.UsableFlashLines.Knowledge, profile.Storage.UsableFlashLines.Value == nil)
	fmt.Println(profile.HardwareVerified, profile.Qualification[0].Status)
	// Output:
	// 204 205 model-only
	// 32 16
	// unknown true
	// false unverified
}
