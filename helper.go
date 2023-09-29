package blink1

import (
	hid "github.com/b1ug/gid"
)

// methods in this file serve as public helper functions

// IsBlink1Device returns true if the device info is about a blink(1) device.
func IsBlink1Device(di *hid.DeviceInfo) bool {
	if di == nil {
		return false
	}
	if di.VendorID == b1VendorID && di.ProductID == b1ProductID {
		return true
	}
	return false
}

// HSBToRGB converts HSB to 8-bit RGB values. The hue is in degrees [0, 360], saturation and brightness/value are percent in the range [0, 100].
// Values outside of these ranges will be clamped.
func HSBToRGB(hue, saturation, brightness float64) (r, g, b uint8) {
	hue = clampFloat64(hue, 0, 360)
	saturation = clampFloat64(saturation, 0, 100)
	brightness = clampFloat64(brightness, 0, 100)
	return convHSBToRGB(hue, saturation/100, brightness/100)
}
