// Package blink1 provides a low-level and high-level interface to the blink(1) USB RGB LED notification light.
package blink1

import (
	"fmt"
	"image/color"
	"math"
	"time"
)

const (
	b1VendorID  = 0x27B8
	b1ProductID = 0x01ED
	cmdBufSize  = 9                 // for mk1&mk2
	cmdBuf3Size = 61                // for mk3+
	reportID    = byte(0x01)        // for normal cmd
	report3ID   = byte(0x02)        // for mk3+
	maxPattern  = uint(12)          // for mk1
	maxPattern2 = uint(32)          // for mk2+
	maxFadeMsec = uint(0xffff * 10) // 10 min 55 sec 350 msec
	maxRepeat   = uint(0xff)        // 255

	minTimeout  = 10 * time.Millisecond // minimum timeout, otherwise the device will treat it as no timeout
	opsInterval = 20 * time.Millisecond // interval between operations
)

var (
	// common values
	durZero  time.Duration
	colorOff = color.RGBA{0x00, 0x00, 0x00, 0x00}
	colorOn  = color.RGBA{0xff, 0xff, 0xff, 0xff}
)

// getMaxPattern returns max pattern number for the generation.
func getMaxPattern(gen uint16) uint {
	if gen >= 2 {
		return maxPattern2
	}
	return maxPattern
}

// convHSBToRGB converts HSB to 8-bit RGB values. The hue is in degrees (0-360), saturation and brightness/value are percent in the range [0, 1].
func convHSBToRGB(h, s, v float64) (r, g, b uint8) {
	h = math.Mod(h, 360)
	h /= 60

	i := math.Floor(h)
	f := h - i
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	switch int(i) % 6 {
	case 0:
		r, g, b = uint8(v*255), uint8(t*255), uint8(p*255)
	case 1:
		r, g, b = uint8(q*255), uint8(v*255), uint8(p*255)
	case 2:
		r, g, b = uint8(p*255), uint8(v*255), uint8(t*255)
	case 3:
		r, g, b = uint8(p*255), uint8(q*255), uint8(v*255)
	case 4:
		r, g, b = uint8(t*255), uint8(p*255), uint8(v*255)
	default: // case 5:
		r, g, b = uint8(v*255), uint8(p*255), uint8(q*255)
	}
	return
}

// convColorToRGB converts color.Color to 8-bit RGB values.
func convColorToRGB(c color.Color) (r, g, b uint8) {
	rr, gg, bb, _ := c.RGBA()
	return uint8(rr >> 8), uint8(gg >> 8), uint8(bb >> 8)
}

// convRGBToColor converts 8-bit RGB values to color.Color.
func convRGBToColor(r, g, b uint8) color.Color {
	return color.RGBA{R: r, G: g, B: b, A: 0xff}
}

// convColorToHex converts color.Color to hex string.
func convColorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

// convHexToColor converts hex string to color.Color.
func convHexToColor(s string) color.Color {
	var r, g, b uint8
	fmt.Sscanf(s, "#%02X%02X%02X", &r, &g, &b)
	return color.RGBA{R: r, G: g, B: b, A: 0xff}
}

// convDurationToActual converts time.Duration to actual time.Duration on the device.
func convDurationToActual(dur time.Duration) time.Duration {
	ms := uint(dur.Milliseconds())
	if ms > maxFadeMsec {
		return time.Duration(maxFadeMsec) * time.Millisecond
	}
	return dur.Truncate(minTimeout)
}

// convDurationToFadeMs converts time.Duration to fadetimeMillis in Big-Endian.
func convDurationToFadeMs(dur time.Duration) (th, tl uint8) {
	return convDurMsToFadeMs(uint(dur.Milliseconds()))
	// fadeMs := uint(dur.Milliseconds())
	// if fadeMs > maxFadeMsec {
	// 	fadeMs = maxFadeMsec
	// }
	// fadeMs /= 10
	// return uint8(fadeMs >> 8), uint8(fadeMs & 0xff)
}

// convFadeMsToDuration converts Big-Endian fadetimeMillis to time.Duration.
func convFadeMsToDuration(th, tl uint8) time.Duration {
	return time.Duration(convFadeMsToDurMs(th, tl)) * time.Millisecond
	// fadeMs := (int(th) << 8) | int(tl)
	// return time.Duration(fadeMs*10) * time.Millisecond
}

// convDurMsToFadeMs converts milliseconds to fadetimeMillis in Big-Endian.
func convDurMsToFadeMs(durMs uint) (th, tl uint8) {
	if durMs > maxFadeMsec {
		durMs = maxFadeMsec
	}
	durMs /= 10
	return uint8(durMs >> 8), uint8(durMs & 0xff)
}

// convFadeMsToDurMs converts Big-Endian fadetimeMillis to milliseconds.
func convFadeMsToDurMs(th, tl uint8) uint {
	durMs := (uint(th) << 8) | uint(tl)
	return durMs * 10
}

// convBoolToByte converts bool to byte.
func convBoolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// convByteToBool converts byte to bool.
func convByteToBool(b byte) bool {
	return b != 0
}

// convPlayingToEmoji converts playing state to emoji.
func convPlayingToEmoji(playing bool) string {
	if playing {
		return `▶️`
	}
	return `⏸`
}

// convDeviceLightState converts DeviceLightState to LightState.
func convDeviceLightState(st DeviceLightState) LightState {
	return LightState{
		Color:    convRGBToColor(st.R, st.G, st.B),
		LED:      st.LED,
		FadeTime: time.Duration(st.FadeTimeMsec) * time.Millisecond,
	}
}

// convLightState converts LightState to DeviceLightState.
func convLightState(st LightState) DeviceLightState {
	r, g, b := convColorToRGB(st.Color)
	return DeviceLightState{
		R:            r,
		G:            g,
		B:            b,
		LED:          st.LED,
		FadeTimeMsec: uint(st.FadeTime.Milliseconds()),
	}
}

// Migrated from https://github.com/todbot/blink1-tool/blob/92661e6d731b46d4bf82e2506c105c5fe433b57d/blink1-lib.c#L676-L700
// Original values from http://rgb-123.com/ws2812-color-output/
//     GammaE=255*(res/255).^(1/.45)
var gammaE = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2,
	2, 2, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5,
	6, 6, 6, 7, 7, 7, 8, 8, 8, 9, 9, 9, 10, 10, 11, 11,
	11, 12, 12, 13, 13, 13, 14, 14, 15, 15, 16, 16, 17, 17, 18, 18,
	19, 19, 20, 21, 21, 22, 22, 23, 23, 24, 25, 25, 26, 27, 27, 28,
	29, 29, 30, 31, 31, 32, 33, 34, 34, 35, 36, 37, 37, 38, 39, 40,
	40, 41, 42, 43, 44, 45, 46, 46, 47, 48, 49, 50, 51, 52, 53, 54,
	55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70,
	71, 72, 73, 74, 76, 77, 78, 79, 80, 81, 83, 84, 85, 86, 88, 89,
	90, 91, 93, 94, 95, 96, 98, 99, 100, 102, 103, 104, 106, 107, 109, 110,
	111, 113, 114, 116, 117, 119, 120, 121, 123, 124, 126, 128, 129, 131, 132, 134,
	135, 137, 138, 140, 142, 143, 145, 146, 148, 150, 151, 153, 155, 157, 158, 160,
	162, 163, 165, 167, 169, 170, 172, 174, 176, 178, 179, 181, 183, 185, 187, 189,
	191, 193, 194, 196, 198, 200, 202, 204, 206, 208, 210, 212, 214, 216, 218, 220,
	222, 224, 227, 229, 231, 233, 235, 237, 239, 241, 244, 246, 248, 250, 252, 255}

// degammaRGB operates degamma correction for 8-bit RGB values.
func degammaRGB(r, g, b uint8) (rr, gg, bb uint8) {
	return gammaE[r], gammaE[g], gammaE[b]
}
