package blink1

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// lifecycleHID records operations on an in-memory handle, without enumerating USB.
type lifecycleHID struct {
	mu             sync.Mutex
	events         []string
	closed         bool
	closeCount     int
	afterClose     bool
	writes         chan []byte
	disableStarted chan struct{}
	disableRelease chan struct{}
	disableOnce    sync.Once
}

func newLifecycleController() (*Controller, *lifecycleHID) {
	h := &lifecycleHID{writes: make(chan []byte, 512)}
	return NewController(&Device{gen: 2, dev: h}), h
}

func (h *lifecycleHID) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	h.closeCount++
	h.events = append(h.events, "close")
}

func (h *lifecycleHID) Write(data []byte) error { return h.WriteFeature(data) }

func (h *lifecycleHID) WriteFeature(data []byte) error {
	h.mu.Lock()
	if h.closed {
		h.afterClose = true
		h.mu.Unlock()
		return errors.New("fake handle is closed")
	}
	h.events = append(h.events, "write")
	h.mu.Unlock()
	h.writes <- append([]byte(nil), data...)
	if len(data) > 2 && data[1] == 'D' && data[2] == 0 && h.disableStarted != nil {
		h.disableOnce.Do(func() { close(h.disableStarted) })
		<-h.disableRelease
	}
	return nil
}

func (h *lifecycleHID) Read(data []byte) (int, error) { return h.ReadFeature(data) }

func (h *lifecycleHID) ReadFeature(data []byte) (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		h.afterClose = true
		return 0, errors.New("fake handle is closed")
	}
	h.events = append(h.events, "read")
	return len(data), nil
}

func waitLifecycle(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("tickle worker did not finish")
	}
}

func readLifecycleReport(t *testing.T, h *lifecycleHID) []byte {
	t.Helper()
	select {
	case report := <-h.writes:
		return report
	case <-time.After(5 * time.Second):
		t.Fatal("tickle report was not written")
		return nil
	}
}

func TestAutoTickleStopAndRestart(t *testing.T) {
	c, h := newLifecycleController()
	defer c.Close()
	c.StopAutoTickle() // stopping before start must also be harmless
	for i := 0; i < 3; i++ {
		if err := c.StartAutoTickle(0, 1, true); err != nil {
			t.Fatal(err)
		}
		done := c.tickleDoneCh
		// Replacing an active worker must finish its disable before returning.
		if err := c.StartAutoTickle(1, 2, true); err != nil {
			t.Fatal(err)
		}
		select {
		case <-done:
		default:
			t.Fatal("replacement started before the old worker finished")
		}
		if report := readLifecycleReport(t, h); report[1] != 'D' || report[2] != 0 {
			t.Fatalf("old worker did not disable tickling: %v", report)
		}
		c.StopAutoTickle()
		c.StopAutoTickle() // used to panic by closing the same channel twice
		if report := readLifecycleReport(t, h); report[1] != 'D' || report[2] != 0 {
			t.Fatalf("stopped worker did not disable tickling: %v", report)
		}
	}
}

func TestCloseWaitsForAutoTickleCleanup(t *testing.T) {
	c, h := newLifecycleController()
	h.disableStarted = make(chan struct{})
	h.disableRelease = make(chan struct{})
	if err := c.StartAutoTickle(0, 1, true); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		c.Close()
		close(done)
	}()
	waitLifecycle(t, h.disableStarted)
	h.mu.Lock()
	closedEarly := h.closed
	h.mu.Unlock()
	// Release cleanup before reporting failures so the worker never leaks.
	close(h.disableRelease)
	waitLifecycle(t, done)
	if closedEarly {
		t.Error("HID handle closed before tickle cleanup completed")
	}
	c.Close()
	c.StopAutoTickle()
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closeCount != 1 || h.afterClose || h.events[len(h.events)-1] != "close" {
		t.Errorf("invalid close sequence: count=%d afterClose=%v events=%v", h.closeCount, h.afterClose, h.events)
	}
}

func TestCloseStopsManualTickle(t *testing.T) {
	c, h := newLifecycleController()
	tick, err := c.StartManualTickle(1, 2, 100*time.Millisecond, true)
	if err != nil {
		t.Fatal(err)
	}
	tick <- struct{}{}
	if report := readLifecycleReport(t, h); report[1] != 'D' || report[2] != 1 || report[6] != 1 || report[7] != 3 {
		t.Fatalf("unexpected manual tickle report: %v", report)
	}
	c.Close()
	// The caller retains channel ownership, even after Controller.Close stops it.
	close(tick)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closeCount != 1 || h.afterClose || len(h.events) != 3 || h.events[2] != "close" {
		t.Errorf("manual worker outlived handle: count=%d afterClose=%v events=%v", h.closeCount, h.afterClose, h.events)
	}
}

func TestTickleStartAfterClose(t *testing.T) {
	for _, closeController := range []bool{true, false} {
		c, _ := newLifecycleController()
		if closeController {
			c.Close()
		} else {
			c.GetDevice().Close()
		}
		if err := c.StartAutoTickle(0, 1, true); err != errDeviceClosed {
			t.Errorf("StartAutoTickle after close = %v", err)
		}
		if _, err := c.StartManualTickle(0, 1, time.Second, true); err != errDeviceClosed {
			t.Errorf("StartManualTickle after close = %v", err)
		}
		c.Close()
	}
}

func TestConcurrentTickleLifecycle(t *testing.T) {
	c, h := newLifecycleController()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 12; j++ {
				if err := c.StartAutoTickle(0, 1, true); err != nil && err != errDeviceClosed {
					t.Errorf("StartAutoTickle = %v", err)
				}
				c.StopAutoTickle()
				if j == 6 {
					c.Close()
				}
			}
		}()
	}
	wg.Wait()
	c.Close()
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closeCount != 1 || h.afterClose {
		t.Errorf("concurrent close: count=%d afterClose=%v", h.closeCount, h.afterClose)
	}
}

func TestDeviceCommandsAfterClose(t *testing.T) {
	c, h := newLifecycleController()
	d := c.GetDevice()
	c.Close()
	for name, call := range map[string]func() error{
		"write":        func() error { return d.SetRGBNow(1, 2, 3, LEDAll) },
		"double write": func() error { return d.SetPatternLine(0, DeviceLightState{}) },
		"read":         func() error { _, _, _, err := d.ReadRGB(LED1); return err },
		"delayed read": func() error { _, err := d.Test(); return err },
		"save pattern": d.SavePattern,
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err != errDeviceClosed {
				t.Errorf("command after close = %v, want %v", err, errDeviceClosed)
			}
		})
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.afterClose || len(h.events) != 1 {
		t.Errorf("closed handle was accessed: %v", h.events)
	}
}
