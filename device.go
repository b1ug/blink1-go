package blink1

import (
	"fmt"
	"sync"
	"time"

	hid "github.com/b1ug/gid"
)

// Device represents a blink(1) device and provides low-level APIs using HID commands for direct control.
type Device struct {
	// profile
	pn  string // product name
	gen uint16 // generation: 1=mk1, 2=mk2, 3=mk3 etc.
	sn  string // serial number

	// state
	mu   sync.Mutex
	info *hid.DeviceInfo
	dev  hid.Device
}

// OpenDevice opens a blink(1) device which is connected to the system.
func OpenDevice(info *hid.DeviceInfo) (*Device, error) {
	// verify device if it is blink(1)
	if info == nil {
		return nil, fmt.Errorf("nil device info")
	}
	if info.VendorID != b1VendorID || info.ProductID != b1ProductID {
		return nil, fmt.Errorf("device is not blink(1)")
	}

	// open device
	dev, err := info.Open()
	if err != nil {
		return nil, err
	}

	// instance
	b1 := &Device{
		pn:   info.Product,
		gen:  info.VersionNumber,
		sn:   info.SerialNumber,
		info: info,
		dev:  dev,
	}
	return b1, nil
}

func (b1 *Device) String() string {
	return fmt.Sprintf("🚦{dev=%q gen=%d sn=%s}", b1.pn, b1.gen, b1.sn)
}

// GetDeviceInfo returns the HID device info.
func (b1 *Device) GetDeviceInfo() *hid.DeviceInfo {
	return b1.info
}

// GetProductName returns the product name.
func (b1 *Device) GetProductName() string {
	return b1.pn
}

// GetGeneration returns the generation number.
func (b1 *Device) GetGeneration() uint16 {
	return b1.gen
}

// GetSerialNumber returns the serial number.
func (b1 *Device) GetSerialNumber() string {
	return b1.sn
}

// Close closes the device and release the kept resources.
func (b1 *Device) Close() {
	b1.mu.Lock()
	defer b1.mu.Unlock()
	if b1.dev != nil {
		b1.dev.Close()
		b1.dev = nil
	}
}

// write sends the specified buffer as feature report to the device.
func (b1 *Device) write(buf []byte) error {
	b1.mu.Lock()
	defer b1.mu.Unlock()

	// send feature report
	if err := b1.dev.WriteFeature(buf); err != nil {
		return fmt.Errorf("b1: write fail: %w", err)
	}
	return nil
}

// doubleWrite works like write but sends the specified buffers one by one.
func (b1 *Device) doubleWrite(buf1, buf2 []byte) error {
	b1.mu.Lock()
	defer b1.mu.Unlock()

	// send feature reports
	if err := b1.dev.WriteFeature(buf1); err != nil {
		return fmt.Errorf("b1: write buf1 fail: %w", err)
	}
	if err := b1.dev.WriteFeature(buf2); err != nil {
		return fmt.Errorf("b1: write buf2 fail: %w", err)
	}
	return nil
}

// read sends the feature report to the device and gets the response and writes it to the specified buffer.
func (b1 *Device) read(buf []byte) error {
	b1.mu.Lock()
	defer b1.mu.Unlock()

	// send feature report
	_ = b1.dev.WriteFeature(buf)

	// get feature report
	if _, err := b1.dev.ReadFeature(buf); err != nil {
		return fmt.Errorf("b1: read fail: %w", err)
	}
	return nil
}

// delayWrite works like read but waits for the specified milliseconds before reading the response.
func (b1 *Device) delayRead(buf []byte, delayMs int) error {
	b1.mu.Lock()
	defer b1.mu.Unlock()

	// send feature report
	_ = b1.dev.WriteFeature(buf)

	// wait
	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	// get feature report
	if _, err := b1.dev.ReadFeature(buf); err != nil {
		return fmt.Errorf("b1: read fail: %w", err)
	}
	return nil
}
