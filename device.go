package blink1

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	hid "github.com/b1ug/gid"
)

var (
	// common errors
	errNilDeviceInfo  = errors.New("b1: nil device info")
	errNotBlink1      = errors.New("b1: device is not blink(1)")
	errDeviceNotFound = fmt.Errorf("b1: device not found")
	errDeviceClosed   = errors.New("b1: device closed")
)

// Device represents a blink(1) device and provides low-level APIs using HID commands for direct control.
type Device struct {
	// profile
	pn  string // product name
	gen uint16 // generation: 1=mk1, 2=mk2, 3=mk3 etc.
	sn  string // serial number

	// state
	mu   sync.Mutex // mutex lock, only for atomic operations like I/O & close
	info *hid.DeviceInfo
	dev  hid.Device
}

// OpenDevice opens a blink(1) device which is connected to the system.
func OpenDevice(info *hid.DeviceInfo) (*Device, error) {
	// verify device if it is blink(1)
	if info == nil {
		return nil, errNilDeviceInfo
	}
	if !IsBlink1Device(info) {
		return nil, errNotBlink1
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

// isClosed reports whether the underlying HID handle has been released.
func (b1 *Device) isClosed() bool {
	b1.mu.Lock()
	defer b1.mu.Unlock()
	return b1.dev == nil
}

// write sends the specified buffer as feature report to the device.
func (b1 *Device) write(buf []byte) error {
	b1.mu.Lock()
	defer b1.mu.Unlock()
	if b1.dev == nil {
		return errDeviceClosed
	}

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
	if b1.dev == nil {
		return errDeviceClosed
	}

	// send feature reports
	if err := b1.dev.WriteFeature(buf1); err != nil {
		return fmt.Errorf("b1: write buf1 fail: %w", err)
	}
	if err := b1.dev.WriteFeature(buf2); err != nil {
		return fmt.Errorf("b1: write buf2 fail: %w", err)
	}
	return nil
}

// read exchanges a feature request and response. minResponseSize is the number
// of bytes needed for the command's parsed fields, including the ID slot.
func (b1 *Device) read(buf []byte, minResponseSize int) error {
	return b1.delayRead(buf, 0, minResponseSize)
}

// delayRead works like read but waits after a successful write before reading.
func (b1 *Device) delayRead(buf []byte, delayMs, minResponseSize int) error {
	b1.mu.Lock()
	defer b1.mu.Unlock()
	if b1.dev == nil {
		return errDeviceClosed
	}
	if len(buf) == 0 || minResponseSize < 1 || minResponseSize > len(buf) {
		return errors.New("b1: invalid feature response buffer size")
	}

	if err := b1.dev.WriteFeature(buf); err != nil {
		return fmt.Errorf("b1: write request fail: %w", err)
	}

	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	// Only the report ID selects the response. Do not let fields left in the
	// request (such as a pattern position) masquerade as returned device data.
	response := make([]byte, len(buf))
	response[0] = buf[0]
	n, err := b1.dev.ReadFeature(response)
	if err != nil {
		return fmt.Errorf("b1: read fail: %w", err)
	}
	if n < 0 || n > len(response) {
		return fmt.Errorf("b1: invalid feature response length %d for buffer size %d", n, len(response))
	}
	if n < minResponseSize {
		return fmt.Errorf("b1: short feature response: got %d bytes, need at least %d: %w", n, minResponseSize, io.ErrUnexpectedEOF)
	}
	copy(buf, response[:n])
	for i := n; i < len(buf); i++ {
		buf[i] = 0
	}
	return nil
}
