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

If you want to get the state from a natural language query, you can try this:

```go
func main() {
	st, err := b1.ParseStateQuery("Fade the first LED to blue in 1.5 seconds.")
	if err != nil {
		panic(err)
	}
	fmt.Println(st)
}
```

For more examples, please refer to [go.dev](https://pkg.go.dev/github.com/b1ug/blink1-go#pkg-examples).

## Contributing

Your contributions are greatly appreciated! Whether you spot a bug, propose an improvement, or wish to add a new feature, your contributions help make this package and blink(1) device better.

## License

[![License: MIT](https://img.shields.io/:license-MIT-blue.svg)](http://opensource.org/licenses/MIT)

As an open-source project, the `blink1-go` package is licensed under the MIT license, making it free and accessible for all.
