# blink1-go

[![Go Reference](https://pkg.go.dev/badge/github.com/b1ug/blink1-go.svg)](https://pkg.go.dev/github.com/b1ug/blink1-go)
[![GitHub Actions](https://github.com/b1ug/blink1-go/actions/workflows/build.yml/badge.svg)](https://github.com/b1ug/blink1-go/actions/workflows/build.yml)

> Yet another Go SDK for blink(1) USB RGB LED notification devices but tightened to perfection

Welcome to `blink1-go`, your Go SDK destination for blink(1) USB RGB LED notification devices delivering an enhanced user experience and extended functionalities.

Far from just another SDK, `blink1-go` was conceived and designed to overcome the limitations and challenges faced by the currently available SDK packages like [GoBlink](https://github.com/todbot/blink1/tree/main/go/GoBlink), [go-blink1](https://github.com/hink/go-blink1). Equipped with a fully implemented feature set, streamlined dependency management, and a user-friendly approach, `blink1-go` enables users with enhanced customization and simplicity, accommodating a wide range of use-cases.

## Why Another blink(1) SDK?

1. **Simplified Dependencies:** Avoiding the convoluted dependency packages and external USB HID libraries prevalent in the existing alternatives, `b1ug/blink1-go` adopts the [`b1ug/gid`](https://github.com/b1ug/gid) package. It uses native APIs to orchestrate USB HID operations on macOS and Windows, leverages [`libusb 1.0+`](https://github.com/libusb/libusb) on Linux.

2. **Complete HID Command Suite:** Breaking away from the partial HID command implementations, `blink1-go` ensures the implementation of all blink(1) mk2 HID commands. Users can easily apply these commands leveraging the *Device API* incorporated in the `blink1-go` package.

3. **Strategically Crafted APIs:** Uniquely constituted, the well-thought *Controller API* provides an intuitive interface for pattern and state management, making the SDK adaptable to a wide range of use cases, broadening the horizon of usage possibilities.

4. **Color Helpers:** `blink1-go` includes helpers offering predefined colors and color-conversion functionalities, enabling users to easily customize their device colors.

5. **Sophisticated Schemata:** It boasts highly refined schemata for pattern and state representation. Coupled with effective serialization and deserialization methods, it ensures seamless compatibility with both JSON and natural language (English), thus offering a more human-readable and comfortable programming experience.

## Installation

Installation remains a simple and very straightforward process with the `go get` command:

```bash
go get -u github.com/b1ug/blink1-go@latest
```

Please note that Go version 1.13 or higher and Cgo is required.

## Continuous integration

The native CI matrix builds the SDK and runs its software tests with the race
detector and `go vet` on Ubuntu 24.04 (amd64/arm64), macOS 15 (arm64/Intel), and
Windows Server 2022 (amd64), using Go 1.26.x and 1.27.x. Go 1.13.15 source
compatibility is checked separately on Ubuntu 22.04. These are configured
checks; a result applies only to the exact commit that ran successfully.

The default tests cover color conversion, parsing, and serialization without
opening hardware. Hardware examples without an `Output` comment are compiled
but are not executed by `go test`. CI does not enumerate or open USB devices
and does not verify physical unplug/replug, older operating systems, Windows
386, or blink(1) mk3. Passing the Go 1.13 source check is not a claim that every
current native dependency or OS supports that old toolchain.

Actions are pinned to reviewed release commits, checkout credentials are not
persisted, and jobs use read-only repository permissions. Module resolution
uses the committed GID version with `GOWORK=off` and `-mod=readonly`; it does not
use a local workspace replacement or upgrade dependencies.

## Usage

Here's how to use the *Device API* to instantly set all LEDs to blue:

```go
import (
	b1 "github.com/b1ug/blink1-go"
)

func main() {
	d, err := b1.OpenNextDevice()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	if err := d.SetRGBNow(0, 0, 0xff, b1.LEDAll); err != nil {
		panic(err)
	}
}
```

And here's how to use the *Controller API* to fade LED 1 to blue and LED 2 to red over 1 second:

```go
func main() {
	c, err := b1.OpenNextController()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	s1 := b1.NewLightState(b1.ColorBlue, 1*time.Second, b1.LED1)
	s2 := b1.NewLightState(b1.ColorRed, 1*time.Second, b1.LED2)
	if err := c.PlayState(s1); err != nil {
		panic(err)
	}
	if err := c.PlayState(s2); err != nil {
		panic(err)
	}
}
```

To extract one light state from an English field query, you can try this:

```go
func main() {
	st, err := b1.ParseStateQuery("Fade the first LED to blue in 1.5 seconds.")
	if err != nil {
		panic(err)
	}
	fmt.Println(st)
}
```

`ParseStateQuery` extracts a color, fade time and LED selector; all three are
required. Parsing does not open a device or change the light. This helper does
not interpret action verbs, negation or multiple instructions, so arbitrary
natural-language instructions should not be treated as executable commands.

Explicit numeric times take precedence over immediate words such as `now`.
Repeated numeric times must agree (`1s` and `1000ms` agree); conflicting,
negative or overflowing times return an error. A standalone `0` or `time=0`
means immediate, while zero inside `led 0` or an RGB/HSB color does not supply a
missing fade time. Fractional milliseconds retain the historical truncation to
whole milliseconds. Existing color names and RGB/HSB conversion semantics are
unchanged.

For more examples, please refer to [go.dev](https://pkg.go.dev/github.com/b1ug/blink1-go#pkg-examples).

## Saving device patterns

`Device.SavePattern` and `Controller.WritePattern` return feature-write errors,
including transport timeouts, instead of treating them as successful saves.
They send the existing save report once, without automatic retries or readback.
A failed write may already have reached the device, so an error leaves flash
persistence uncertain; do not automatically repeat the save. A nil error means
the transport API reported success, not that persistence across power loss was
verified.

This intentionally changes earlier SDK error handling. The upstream C
[`blink1_savePattern` implementation](https://github.com/todbot/blink1-tool/blob/main/blink1-lib.c)
assumes success because flash programming can time out USB. That compatibility
workaround explains the previous behavior; it is not evidence that persistence
was confirmed. This SDK now exposes the communication error without declaring
whether the device applied the save.

## Contributing

Your contributions are greatly appreciated! Whether you spot a bug, propose an improvement, or wish to add a new feature, your contributions help make this package and blink(1) device better.

## License

[![License: MIT](https://img.shields.io/:license-MIT-blue.svg)](http://opensource.org/licenses/MIT)

As an open-source project, the `blink1-go` package is licensed under the MIT license, making it free and accessible for all.
