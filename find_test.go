package blink1

import (
	"errors"
	"reflect"
	"testing"
	"time"

	hid "github.com/b1ug/gid"
)

func fakeDeviceInfoFinder(t *testing.T, infos []*hid.DeviceInfo) (func(uint16, uint16) <-chan *hid.DeviceInfo, <-chan *hid.DeviceInfo) {
	t.Helper()
	channel := make(chan *hid.DeviceInfo, len(infos))
	for _, info := range infos {
		channel <- info
	}
	close(channel)
	return func(vendor, product uint16) <-chan *hid.DeviceInfo {
		if vendor != 0x27b8 || product != 0x01ed {
			t.Fatalf("discovery was not restricted to blink(1): %04x:%04x", vendor, product)
		}
		return channel
	}, channel
}

func TestListDeviceInfoUsesExactIDsAndStableSerialOrder(t *testing.T) {
	first := &hid.DeviceInfo{VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "20000002", Path: "first-interface"}
	second := &hid.DeviceInfo{VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "20000002", Path: "second-interface"}
	lower := &hid.DeviceInfo{VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "10000001"}
	// Defensive metadata filtering remains in the SDK, including nil entries.
	find, channel := fakeDeviceInfoFinder(t, []*hid.DeviceInfo{first, nil, {VendorID: 0x1234, ProductID: 0x01ed}, second, lower, {VendorID: 0x27b8, ProductID: 0x9999}})
	if got, want := listDeviceInfo(find), []*hid.DeviceInfo{lower, first, second}; !reflect.DeepEqual(got, want) {
		t.Fatalf("stable blink-only enumeration: got=%v want=%v", got, want)
	}
	if len(channel) != 0 {
		t.Fatal("enumeration channel was not fully consumed")
	}
}

func TestListDeviceInfoPreservesNilWhenEmpty(t *testing.T) {
	find, channel := fakeDeviceInfoFinder(t, nil)
	if got := listDeviceInfo(find); got != nil || len(channel) != 0 {
		t.Fatalf("empty discovery changed nil-slice semantics: %#v", got)
	}
}

func TestFindDeviceInfoBySerialNumberKeepsFirstExactMatchAndConsumesAll(t *testing.T) {
	first := &hid.DeviceInfo{VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "ABCDEF01", Path: "first"}
	duplicate := &hid.DeviceInfo{VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "ABCDEF01", Path: "duplicate"}
	find, channel := fakeDeviceInfoFinder(t, []*hid.DeviceInfo{first, {VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "abcdef01"}, duplicate, {VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "FFFFFFFF"}})
	got, err := findDeviceInfoBySerialNumber("ABCDEF01", find)
	if err != nil || got != first {
		t.Fatalf("exact first-match behavior changed: got=%v err=%v", got, err)
	}
	if len(channel) != 0 {
		t.Fatal("returning a match left enumeration items undrained")
	}
}

func TestFindDeviceInfoBySerialNumberPreservesNotFoundError(t *testing.T) {
	for _, serial := range []string{"missing", "ABCDEF01", ""} {
		find, channel := fakeDeviceInfoFinder(t, []*hid.DeviceInfo{{VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "abcdef01"}, {VendorID: 0x1234, ProductID: 0x01ed, SerialNumber: serial}})
		got, err := findDeviceInfoBySerialNumber(serial, find)
		if got != nil || !errors.Is(err, errDeviceNotFound) {
			t.Fatalf("missing serial %q: got=%v err=%v", serial, got, err)
		}
		want := "b1: device not found for \"" + serial + "\""
		if err.Error() != want || len(channel) != 0 {
			t.Fatalf("not-found error or channel changed: err=%v want=%q remaining=%d", err, want, len(channel))
		}
	}
}

func TestFindDeviceInfoBySerialNumberReleasesUnbufferedProducer(t *testing.T) {
	first := &hid.DeviceInfo{VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "20000001"}
	last := &hid.DeviceInfo{VendorID: 0x27b8, ProductID: 0x01ed, SerialNumber: "20000002"}
	producerDone := make(chan struct{})
	abort := make(chan struct{})
	defer close(abort)
	find := func(vendor, product uint16) <-chan *hid.DeviceInfo {
		if vendor != 0x27b8 || product != 0x01ed {
			t.Fatal("unbuffered discovery lost the exact blink(1) filter")
		}
		channel := make(chan *hid.DeviceInfo)
		go func() {
			defer close(channel)
			for _, info := range []*hid.DeviceInfo{first, last} {
				select {
				case channel <- info:
				case <-abort:
					return
				}
			}
			close(producerDone)
		}()
		return channel
	}
	got, err := findDeviceInfoBySerialNumber(first.SerialNumber, find)
	if err != nil || got != first {
		t.Fatalf("first exact match changed: got=%v err=%v", got, err)
	}
	select {
	case <-producerDone:
	case <-time.After(time.Second):
		t.Fatal("matching early left an unbuffered enumeration producer blocked")
	}
}
