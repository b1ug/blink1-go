package blink1

import (
	"errors"
	"image/color"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/b1ug/colorkit"
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

func TestInvalidTypedStatesDoNotWrite(t *testing.T) {
	var typedNil *color.RGBA
	for _, tc := range []struct {
		name  string
		state LightState
		field string
	}{
		{"missing color", LightState{}, "color"},
		{"typed nil", LightState{Color: typedNil}, "color"},
		{"LED", LightState{Color: ColorRed, LED: LEDIndex(3)}, "led"},
		{"negative duration", LightState{Color: ColorRed, FadeTime: -time.Millisecond}, "fade"},
		{"sub-unit fade", LightState{Color: ColorRed, FadeTime: 15 * time.Millisecond}, "fade"},
		{"fractional ms", LightState{Color: ColorRed, FadeTime: time.Nanosecond}, "fade"},
		{"overflow", LightState{Color: ColorRed, FadeTime: 655360 * time.Millisecond}, "fade"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			controller, handle := newLifecycleController()
			defer controller.Close()
			for _, play := range []func(LightState) error{controller.PlayState, controller.PlayStateBlocking} {
				var detail *ParseError
				if err := play(tc.state); !errors.As(err, &detail) || detail.Field != tc.field {
					t.Fatalf("missing field diagnostic: %v", err)
				}
			}
			select {
			case report := <-handle.writes:
				t.Fatalf("invalid state wrote a device report: %v", report)
			default:
			}
		})
	}
}

func TestDirectRGB8StateUsesExistingControllerPath(t *testing.T) {
	controller, handle := newLifecycleController()
	defer controller.Close()
	light := colorkit.RGB8{R: 0x00, G: 0x00, B: 0xff}
	// No channel extraction or color.RGBA conversion is required by this caller.
	if err := controller.PlayState(NewLightState(light, 20*time.Millisecond, LED1)); err != nil {
		t.Fatal(err)
	}
	if got, want := readLifecycleReport(t, handle), []byte{1, 'c', 0, 0, 255, 0, 2, 1, 0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("wrong protocol report: got %v want %v", got, want)
	}
}

func TestDirectRGB8ColorKeepsGammaAndRawRGBRemainsDistinct(t *testing.T) {
	controller, handle := newLifecycleController()
	defer controller.Close()
	light := colorkit.RGB8{R: 128, G: 64, B: 32}
	if err := controller.PlayColor(light); err != nil {
		t.Fatal(err)
	}
	direct := readLifecycleReport(t, handle)
	if err := controller.PlayColor(color.RGBA{R: 128, G: 64, B: 32, A: 255}); err != nil {
		t.Fatal(err)
	}
	existing := readLifecycleReport(t, handle)
	if !reflect.DeepEqual(direct, existing) {
		t.Fatalf("direct color changed the established gamma path: %v != %v", direct, existing)
	}
	if err := controller.PlayRGB(128, 64, 32); err != nil {
		t.Fatal(err)
	}
	raw := readLifecycleReport(t, handle)
	if !reflect.DeepEqual(raw[2:5], []byte{128, 64, 32}) || reflect.DeepEqual(direct[2:5], raw[2:5]) {
		t.Fatalf("expected raw channels to remain distinct from gamma output: direct=%v raw=%v", direct, raw)
	}
	controller.SetGammaCorrection(false)
	if err := controller.PlayColor(light); err != nil {
		t.Fatal(err)
	}
	if withoutGamma := readLifecycleReport(t, handle); !reflect.DeepEqual(withoutGamma, raw) {
		t.Fatalf("disabled gamma should preserve raw channels: %v != %v", withoutGamma, raw)
	}
}

func TestNilDirectColorsDoNotWrite(t *testing.T) {
	controller, handle := newLifecycleController()
	defer controller.Close()
	var typedNil *color.RGBA
	for _, value := range []color.Color{nil, typedNil} {
		if err := controller.PlayColor(value); err == nil {
			t.Fatal("missing color was accepted")
		}
	}
	select {
	case report := <-handle.writes:
		t.Fatalf("missing color wrote to the device: %v", report)
	default:
	}
}
