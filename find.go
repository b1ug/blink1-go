package blink1

import (
	"sort"
	"sync"

	hid "github.com/b1ug/gid"
)

var (
	devInfoMu   sync.RWMutex
	devInfoList []*hid.DeviceInfo
)

// ListDeviceInfo returns all HID device info of all blink(1) devices which are connected to the system. The returned slice is sorted by serial number.
func ListDeviceInfo() []*hid.DeviceInfo {
	infos := hid.ListAllDevices(IsBlink1Device)
	sort.SliceStable(infos, func(i, j int) bool {
		return infos[i].SerialNumber < infos[j].SerialNumber
	})
	return infos
}

// FindDeviceInfoBySerialNumber finds a connected blink(1) device with serial number and returns its HID device info.
func FindDeviceInfoBySerialNumber(sn string) (*hid.DeviceInfo, error) {
	dev := hid.ListFirstDevice(func(di *hid.DeviceInfo) bool {
		return IsBlink1Device(di) && di.SerialNumber == sn
	})
	// not found
	if dev == nil {
		return nil, errDeviceNotFound
	}
	return dev, nil
}

// FindNextDeviceInfo returns the next HID device info of a blink(1) device which is connected to the system.
func FindNextDeviceInfo() (di *hid.DeviceInfo, err error) {
	devInfoMu.Lock()
	defer devInfoMu.Unlock()

	// nil list: need init or reset
	if devInfoList == nil {
		devInfoList = ListDeviceInfo()
	}

	// empty list: last device already returned
	if len(devInfoList) == 0 {
		devInfoList = nil
		return nil, errDeviceNotFound
	}
	di = devInfoList[0]
	devInfoList = devInfoList[1:]
	return di, nil
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

// OpenDeviceBySerialNumber finds a connected blink(1) device with serial number and opens it as device.
func OpenDeviceBySerialNumber(sn string) (*Device, error) {
	// find
	di, err := FindDeviceInfoBySerialNumber(sn)
	if err != nil {
		return nil, err
	}

	// open
	return OpenDevice(di)
}

// OpenControllerBySerialNumber finds a connected blink(1) device with serial number and opens it as controller.
func OpenControllerBySerialNumber(sn string) (*Controller, error) {
	// find
	di, err := FindDeviceInfoBySerialNumber(sn)
	if err != nil {
		return nil, err
	}

	// open
	return OpenController(di)
}
