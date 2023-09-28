package blink1

import (
	"fmt"
	"sync"

	hid "github.com/b1ug/gid"
)

var (
	devInfoMu sync.RWMutex
	devInfoCh <-chan *hid.DeviceInfo

	// common errors
	errDeviceNotFound = fmt.Errorf("b1: device not found")
)

// FindNextDeviceInfo returns the next HID device info of a blink(1) device which is connected to the system.
func FindNextDeviceInfo() (di *hid.DeviceInfo, err error) {
	devInfoMu.Lock()
	defer devInfoMu.Unlock()

	// init or reset
	if devInfoCh == nil {
		devInfoCh = hid.Devices()
	}

	// loop until found or end
	for {
		// get next
		di, ok := <-devInfoCh
		if !ok {
			devInfoCh = nil
			return nil, errDeviceNotFound
		}

		// check
		if di.VendorID == b1VendorID && di.ProductID == b1ProductID {
			return di, nil
		}
	}
}

// OpenNextDevice opens the next blink(1) device which is connected to the system and returns as device.
func OpenNextDevice() (*Device, error) {
	// find
	di, err := FindNextDeviceInfo()
	if err != nil {
		return nil, err
	}

	// open
	return OpenDevice(di)
}

// OpenNextController opens the next blink(1) device which is connected to the system and returns as controller.
func OpenNextController() (*Controller, error) {
	// find
	di, err := FindNextDeviceInfo()
	if err != nil {
		return nil, err
	}

	// open
	return OpenController(di)
}

// ListDeviceInfo returns all HID device info of all blink(1) devices which are connected to the system.
func ListDeviceInfo() []*hid.DeviceInfo {
	var infos []*hid.DeviceInfo
	for di := range hid.Devices() {
		if di.VendorID == b1VendorID && di.ProductID == b1ProductID {
			infos = append(infos, di)
		}
	}
	return infos
}
