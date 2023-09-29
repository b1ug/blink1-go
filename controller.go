package blink1

import (
	"fmt"
	"sync"

	hid "github.com/b1ug/gid"
)

// Controller provides a high-level API for operating blink(1) devices, abstracting away the low-level details.
type Controller struct {
	mu     sync.Mutex
	dev    *Device
	quitCh chan struct{}
}

// OpenController opens a blink(1) controller for device which is connected to the system.
func OpenController(info *hid.DeviceInfo) (*Controller, error) {
	dev, err := OpenDevice(info)
	if err != nil {
		return nil, err
	}
	return &Controller{dev: dev}, nil
}

// NewController creates a blink(1) controller for existing device instance.
func NewController(dev *Device) *Controller {
	return &Controller{dev: dev}
}

func (c *Controller) String() string {
	return fmt.Sprintf("🎮(ctrl=%q gen=%d sn=%s)", c.dev.pn, c.dev.gen, c.dev.sn)
}

// GetDevice returns the underlying blink(1) device.
func (c *Controller) GetDevice() *Device {
	return c.dev
}

// Close closes the device and release the kept resources.
func (c *Controller) Close() {
	c.dev.Close()
}
