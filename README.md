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

5. **Typed States and Patterns:** State and pattern values support explicit text/JSON codecs. A documented Set/Off/Fade grammar returns a ready-to-use `LightState` without guessing general natural-language intent.

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
uses the committed GID and colorkit versions with `GOWORK=off` and
`-mod=readonly`; it does not use a local workspace replacement or upgrade
dependencies.

## Hexadecimal color helpers

`HexToColor` and `RGBToHex` use the pure-Go
[`colorkit v0.1.0`](https://github.com/b1ug/colorkit/tree/v0.1.0) hexadecimal core.
`HexToColor` accepts three or six hex digits, with an optional `#`, mixed case,
and surrounding ASCII whitespace. The original input is limited to 64 ASCII
bytes and must contain one complete literal. Internal whitespace, trailing
non-hex content, alpha, and color names are rejected. This fixes cases such as
`"12 345"` that the earlier formatted scanner could consume as three channels.
Errors return a nil color, match `colorkit.ErrSyntax` through `errors.Is`, and
expose `*colorkit.Error` through `errors.As`; their text is not a stable contract.

Successful results remain concrete `color.RGBA` values with full opacity.
`RGBToHex`, the existing state text format, and SDK hexadecimal output retain
uppercase digits. Complete command/color parsing uses the same colorkit values
and adapts them inside the SDK, without asking callers to convert types.
See [single-state parsing](docs/state-parser.md) for grammar, defaults,
diagnostics, and the documented CSS color-name migration.

## Usage

### Pass color values directly

`colorkit.RGB8` implements Go's `image/color.Color` interface, which is already
accepted by `NewLightState` and `Controller.PlayColor`. Pass the value directly;
there is no need to unpack R/G/B or construct another color type.

```go
import (
	"fmt"
	"time"

	b1 "github.com/b1ug/blink1-go"
	"github.com/b1ug/colorkit"
)

light, err := colorkit.Parse("#abc")
if err != nil {
	return err
}
state := b1.NewLightState(light, 250*time.Millisecond, b1.LED1)
fmt.Println(state) // 🎨(color=#AABBCC led=1 fade=250ms)
```

This fragment belongs in a function that returns an error. With an already
opened controller, `controller.PlayColor(light)` accepts the same value.
To create it from HSV, use:

```go
light, err := (colorkit.HSV{H: 210, S: 75, V: 90}).RGB8()
```

Handle the error, then pass `light` directly to either SDK API. H uses degrees;
S and V use percentages. For hexadecimal input through the existing SDK API,
`light, err := b1.HexToColor("abc")` is also sufficient.

`PlayColor` and `PlayState` apply the controller's Gamma setting, enabled by
default. `PlayRGB` sends raw channels and bypasses Gamma; unpacking RGB8 into
`PlayRGB` is not equivalent to passing the color directly.

The runnable examples verify value conversion and `LightState` formatting
without hardware. The controller example is compiled but has no `Output`
comment, so `go test` does not execute its device operations. Calling
`colorkit.Parse`, `ParseColor`, and `GetColorByName` share the CSS name table.
The four historical name changes and five retained Go-only constants are
listed in [the migration notes](docs/state-parser.md#compact-data-and-behavior-changes).

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

### Parse a complete command into a ready-to-use state

The SDK handles the translation; there is no intermediate state DTO to copy:

```go
func main() {
	st, err := b1.ParseStateQuery("fade to blue over 1.5s on led 1")
	if err != nil {
		panic(err)
	}
	fmt.Println(st)
}
```

`set red` and `off` use immediate output and default to LED 0. Fade requires
an explicit duration; `led=1 color=red fade=1s` is the exact three-key form.
All meaningful input must be consumed: unknown text, negation, duplicate fields,
and unsupported effects return `*ParseError` and a zero state. Parsed durations
must use exact 10ms units up to 655350ms; they are not rounded or clipped.

The returned state goes directly to `controller.PlayState(st)`. Parsing does not
open a device or change the light. Parse before opening hardware; keep a
controller open when processing multiple commands. Higher-level effects and
holds belong to nb1's Action/Sequence layer, not a single `LightState`.
See [the full grammar and error contract](docs/state-parser.md).

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
