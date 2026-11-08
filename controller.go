package blink1

import (
	"fmt"
	"sync"

	hid "github.com/b1ug/gid"
)

// Controller provides a high-level API for operating blink(1) devices, abstracting away the low-level details.
type Controller struct {
	mu            sync.Mutex
	dev           *Device
	gamma         bool
	quitCh        chan struct{}
	tickleDoneCh  chan struct{}
	manualQuitCh  chan struct{}
	manualTickles sync.WaitGroup
	closed        bool
}

// OpenController opens a blink(1) controller for device which is connected to the system.
func OpenController(info *hid.DeviceInfo) (*Controller, error) {
	dev, err := OpenDevice(info)
	if err != nil {
		return nil, err
	}
	return &Controller{dev: dev, gamma: true}, nil
}

// NewController creates a blink(1) controller for existing device instance.
func NewController(dev *Device) *Controller {
	return &Controller{dev: dev, gamma: true}
}

func (c *Controller) String() string {
	return fmt.Sprintf("🎮(ctrl=%q gen=%d sn=%s)", c.dev.pn, c.dev.gen, c.dev.sn)
}

// GetDevice returns the underlying blink(1) device.
func (c *Controller) GetDevice() *Device {
	return c.dev
}

// Close stops all tickle workers, waits for them to finish, and closes the device.
// It is safe to call Close more than once.
func (c *Controller) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	c.stopAutoTickle()
	if c.manualQuitCh != nil {
		close(c.manualQuitCh)
	}
	c.manualTickles.Wait()
	if c.dev != nil {
		c.dev.Close()
	}
}

// SetGammaCorrection sets the gamma correction on/off for the controller. Default is on.
// If it is true, the gamma correction will be applied for state and pattern while playing or writing.
func (c *Controller) SetGammaCorrection(on bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gamma = on
}
