package blink1

import (
	"errors"
	"fmt"
	"image/color"
	"time"
)

var (
	errInvalidPosition    = errors.New("b1: invalid pattern position")
	errInvalidRepeatTimes = errors.New("b1: invalid pattern repeat times")
	errInvalidTimeout     = errors.New("b1: invalid timeout")
)

// GetFirmwareVersion returns the firmware version of the device.
func (c *Controller) GetFirmwareVersion() (int, error) {
	return c.dev.GetVersion()
}

// PlayStateBlocking fades the given LED to the specified RGB color over the specified time, and blocks until the fade is finished.
func (c *Controller) PlayStateBlocking(st LightState) error {
	// play state
	if err := c.PlayState(st); err != nil {
		return err
	}

	// block until fade is finished
	if dur := convDurationToActual(st.FadeTime); dur > 0 {
		time.Sleep(dur)
	}
	return nil
}

// PlayState fades the given LED to the specified RGB color over the specified time.
func (c *Controller) PlayState(st LightState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	r, g, b := degammaRGB(convColorToRGB(st.Color))
	msec := uint(st.FadeTime.Milliseconds())
	return c.dev.FadeToRGB(r, g, b, msec, st.LED)
}

// PlayColor fades the all LED to the specified RGB color immediately.
func (c *Controller) PlayColor(cl color.Color) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	r, g, b := degammaRGB(convColorToRGB(cl))
	return c.dev.SetRGBNow(r, g, b, LEDAll)
}

// PlayRGB fades the all LED to the specified RGB color immediately.
func (c *Controller) PlayRGB(r, g, b byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.dev.SetRGBNow(r, g, b, LEDAll)
}

// PlayHSB fades the all LED to the specified HSB/HSV color immediately.
// Valid hue range is [0, 360], saturation range and brightness/value range is [0, 100].
// Values outside of the valid range will be clamped to the range.
func (c *Controller) PlayHSB(hue, saturation, brightness float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	r, g, b := degammaRGB(convHSBToRGB(hue, saturation, brightness))
	return c.dev.SetRGBNow(r, g, b, LEDAll)
}

// ReadColor reads the current color of the specified LED.
func (c *Controller) ReadColor(ledN LEDIndex) (color.Color, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	r, g, b, err := c.dev.ReadRGB(ledN)
	if err != nil {
		return nil, fmt.Errorf("b1: failed to read rgb: %w", err)
	}
	return convRGBToColor(r, g, b), nil
}

// PlayPatternBlocking plays the given pattern, and blocks until the pattern is finished. It may block forever if the pattern is set to loop forever.
// If the pattern has no states, it will only play the pattern without writing states to the device's RAM, and blocks until the pattern is finished.
func (c *Controller) PlayPatternBlocking(pt Pattern) error {
	// play pattern
	if err := c.PlayPattern(pt); err != nil {
		return err
	}

	// block until pattern is finished
	if pt.RepeatTimes == 0 {
		// infinite loop, block forever
		<-make(chan struct{})
	} else {
		// otherwise read pattern to get total duration
		startPos, endPos := pt.StartPosition, pt.EndPosition
		if endPos == 0 {
			endPos = getMaxPattern(c.dev.gen) - 1
		}
		// read pattern to get total duration
		var totalDur time.Duration
		for i := startPos; i <= endPos; i++ {
			var st DeviceLightState
			if err := retryWorkload(func() (ie error) {
				st, ie = c.dev.ReadPatternLine(i)
				return ie
			}); err == nil {
				totalDur += time.Duration(st.FadeTimeMsec) * time.Millisecond
			} else {
				return fmt.Errorf("b1: failed to read pattern line %d: %w", i, err)
			}
		}
		// sleep for total duration
		time.Sleep(totalDur * time.Duration(pt.RepeatTimes))
	}
	return nil
}

// PlayPattern plays the given pattern. If the pattern has no states, it will only play the pattern without writing states to the device's RAM
func (c *Controller) PlayPattern(pt Pattern) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ensure range is valid
	if !c.isPosRangeValid(pt.StartPosition, pt.EndPosition) {
		return errInvalidPosition
	}
	if pt.RepeatTimes > maxRepeat {
		return errInvalidRepeatTimes
	}
	if pt.EndPosition == 0 {
		pt.EndPosition = getMaxPattern(c.dev.gen) - 1
	}

	// load pattern to RAM
	if err := c.LoadPattern(pt.StartPosition, pt.EndPosition, pt.States); err != nil {
		return err
	}

	// play pattern
	return c.dev.PlayLoop(true, pt.StartPosition, pt.EndPosition, pt.RepeatTimes)
}

// LoadPattern loads the given pattern to the device's RAM.
func (c *Controller) LoadPattern(posStart, posEnd uint, states []LightState) error {
	sc := len(states) // sc for state counter
	if sc == 0 {
		// no states, just do nothing
		return nil
	}
	if !c.isPosRangeValid(posStart, posEnd) {
		// ensure range is valid
		return errInvalidPosition
	}
	if posEnd == 0 {
		// set posEnd to patt_max-1 if posEnd == 0
		posEnd = getMaxPattern(c.dev.gen) - 1
	}

	// set patterns
	pc := 0 // pc for position counter
	for pos := posStart; pos <= posEnd; pos++ {
		// convert state with degamma and set as pattern
		st := convLightState(states[pc])
		st.R, st.G, st.B = degammaRGB(st.R, st.G, st.B)

		// operate on device
		if err := retryWorkload(func() error {
			return c.dev.SetPatternLine(pos, st)
		}); err != nil {
			return fmt.Errorf("b1: failed to set pattern line %d: %w", pos, err)
		}

		// quit if all left states are filled
		if pc++; pc >= sc {
			break
		}

		// sleep for a little while to avoid hardware errors
		time.Sleep(opsInterval)
	}
	return nil
}

// ReadPattern reads the current pattern in the device's RAM.
func (c *Controller) ReadPattern() ([]LightState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ls []LightState
	for pos, posMax := uint(0), getMaxPattern(c.dev.gen); pos < posMax; pos++ {
		var st DeviceLightState
		if err := retryWorkload(func() (ie error) {
			st, ie = c.dev.ReadPatternLine(pos)
			return ie
		}); err != nil {
			return nil, fmt.Errorf("b1: failed to read pattern line %d: %w", pos, err)
		}
		ls = append(ls, convDeviceLightState(st))
	}
	return ls, nil
}

// WritePattern writes the pattern in the device's RAM to its flash.
func (c *Controller) WritePattern() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.dev.SavePattern()
}

// IsPatternPlaying returns true if the pattern is playing.
func (c *Controller) IsPatternPlaying() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	st, err := c.dev.ReadPlaystate()
	if err != nil {
		return false, fmt.Errorf("b1: failed to read play state: %w", err)
	}
	return st.IsPlaying, nil
}

// GetPatternState returns the current state of the pattern that is playing.
func (c *Controller) GetPatternState() (PatternState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	st, err := c.dev.ReadPlaystate()
	if err != nil {
		return PatternState{}, err
	}
	return PatternState{
		IsPlaying:       st.IsPlaying,
		CurrentPosition: st.CurrentPos,
		StartPosition:   st.LoopStartPos,
		EndPosition:     st.LoopEndPos,
		RepeatTimes:     st.RepeatTimes,
	}, nil
}

// StopPlaying stops playing the pattern and turns off all the LEDs.
// It will stop the current playing patterns, whether it is started by StartPlaying() or StartAuto/ManualTickle(), and turn off all the LEDs.
// If the pattern is not playing, it only turns off all the LEDs.
// It will NOT stop the auto/manual tickle.
func (c *Controller) StopPlaying() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.dev.SetTickleMode(false, false, 0, 0, 0)
}

// StartAutoTickle sets the device to automatically tickle every 2 seconds.
// If the auto tickle is already started, it will be stopped and restarted.
// If keepOld is true, the current pattern will be kept playing, otherwise it will be stopped.
//
// To stop the auto tickle, call StopAutoTickle().
func (c *Controller) StartAutoTickle(posStart, posEnd uint, keepOld bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ensure range is valid
	if !c.isPosRangeValid(posStart, posEnd) {
		return errInvalidPosition
	}

	// if already started, stop it first
	if c.quitCh != nil {
		close(c.quitCh)
	}

	// prepare timeout ticker
	timeout := 2 * time.Second
	timeoutMsec := uint(timeout.Milliseconds())
	timeoutMsec += timeoutMsec >> 1 // add 50% to timeout
	ticker := time.NewTicker(timeout)
	c.quitCh = make(chan struct{})

	// start auto tickle
	go func() {
		for {
			select {
			case <-ticker.C:
				_ = c.dev.SetTickleMode(true, keepOld, posStart, posEnd, timeoutMsec)
			case <-c.quitCh:
				// quit when tickQuit is closed
				ticker.Stop()
				_ = c.dev.SetTickleMode(false, keepOld, 0, 0, 0)
				return
			}
		}
	}()
	return nil
}

// StopAutoTickle stops the device from automatically tickling.
func (c *Controller) StopAutoTickle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quitCh != nil {
		close(c.quitCh)
	}
}

// StartManualTickle sets the device to tickle manually.
// The timeout should be at least 10ms, or it will be ignored by the firmware.
// Signals should be sent to the returned channel to tickle before the timeout, otherwise the given pattern will be played.
// If keepOld is true, the current pattern will be kept playing, otherwise it will be stopped.
//
// To stop the manual tickle, close the returned channel.
func (c *Controller) StartManualTickle(posStart, posEnd uint, timeout time.Duration, keepOld bool) (chan<- struct{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ensure start < end and end < max
	if !c.isPosRangeValid(posStart, posEnd) {
		return nil, errInvalidPosition
	}
	if timeout < minTimeDur {
		return nil, errInvalidTimeout
	}

	// prepare manual ticker
	tickCh := make(chan struct{})
	timeoutMsec := uint(timeout.Milliseconds())

	// start tickle
	go func() {
		for range tickCh {
			_ = c.dev.SetTickleMode(true, keepOld, posStart, posEnd, timeoutMsec)
		}
		// quit when tickCh is closed
		_ = c.dev.SetTickleMode(false, keepOld, 0, 0, 0)
	}()
	return tickCh, nil
}

// isPosRangeValid checks if the given position range is valid.
func (c *Controller) isPosRangeValid(start, end uint) bool {
	// check pattern to ensure start <= end and end < max, 0 is a special case equals to last position
	mp := getMaxPattern(c.dev.gen)
	return (start <= end && end < mp) || (start < mp && end == 0)
}
