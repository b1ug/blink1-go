package blink1

import (
	"fmt"
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
	return listDeviceInfo(hid.FindDevices)
}

// Keep the filter at the enumeration entry, rather than enumerating unrelated
// HID collections and discarding them later. The transport determines how early
// this restriction is applied by each native backend.
func listDeviceInfo(findDevices func(uint16, uint16) <-chan *hid.DeviceInfo) []*hid.DeviceInfo {
	var infos []*hid.DeviceInfo
	for info := range findDevices(b1VendorID, b1ProductID) {
		if IsBlink1Device(info) {
			infos = append(infos, info)
		}
	}
	sort.SliceStable(infos, func(i, j int) bool {
		return infos[i].SerialNumber < infos[j].SerialNumber
	})
	return infos
}

// FindDeviceInfoBySerialNumber finds a connected blink(1) device with serial number and returns its HID device info.
func FindDeviceInfoBySerialNumber(sn string) (*hid.DeviceInfo, error) {
	return findDeviceInfoBySerialNumber(sn, hid.FindDevices)
}

func findDeviceInfoBySerialNumber(sn string, findDevices func(uint16, uint16) <-chan *hid.DeviceInfo) (*hid.DeviceInfo, error) {
	// Collect the finite enumeration fully before returning a match, so neither
	// the channel producer nor its native enumeration cleanup is left waiting.
	for _, info := range listDeviceInfo(findDevices) {
		if info.SerialNumber == sn {
			return info, nil
		}
	}
	return nil, fmt.Errorf("%w for %q", errDeviceNotFound, sn)
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
