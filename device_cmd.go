package blink1

import "fmt"

// FadeToRGB fades the given LED to the specified RGB color over the specified time.
//
// The fadeMsec parameter specifies the fade time in milliseconds, fade time less than 10ms will be processed as no fade time by the firmware.
// The ledN parameter specifies which LED to control: 0=all, 1=top LED, 2=bottom LED.
//
// Returns an error if there was a problem communicating with the device.
func (b1 *Device) FadeToRGB(r, g, b byte, fadeMsec uint, ledN LEDIndex) error {
	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'c'
	buf[2], buf[3], buf[4] = r, g, b
	buf[5], buf[6] = convDurMsToFadeMs(fadeMsec)
	buf[7] = ledN.ToByte()

	// execute
	return b1.write(buf)
}

// SetRGBNow sets the given LED to the specified RGB color immediately.
//
// The ledN parameter specifies which LED to control: 0=all, 1=top LED, 2=bottom LED.
// For mk2+ devices, ledN > 0 will set all LEDs to the white color (255, 255, 255) and ignore the RGB values due to a firmware bug.
//
// Returns an error if there was a problem communicating with the device.
func (b1 *Device) SetRGBNow(r, g, b byte, ledN LEDIndex) error {
	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'n'
	buf[2], buf[3], buf[4] = r, g, b
	buf[7] = ledN.ToByte()

	// execute
	return b1.write(buf)
}

// ReadRGB reads the current RGB color of the specified LED.
//
// The ledN parameter specifies which LED to control: 0=all, 1=top LED, 2=bottom LED.
// For mk2+ devices, ledN == 0 will return the RGB values of the first LED.
//
// Returns the RGB values or an error if there was a problem communicating with the device.
func (b1 *Device) ReadRGB(ledN LEDIndex) (r, g, b byte, err error) {
	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'r'
	buf[7] = ledN.ToByte()

	// execute
	if err = b1.read(buf); err != nil {
		return
	}

	// parse result
	r, g, b = buf[2], buf[3], buf[4]
	return
}

// PlayLoop starts or stops a sub-loop of the color pattern.
//
// The play parameter specifies whether to start or stop the loop.
// The posStart and posEnd parameters specify the start and end positions of the loop, where positions should be [0, patt_max-1]. If posEnd is 0, it will be set to patt_max by the firmware.
// There is no check for posStart < posEnd, and the device will play the loop in even if posStart > posEnd, i.e. the loop will play from posStart to maxPos-1 and then 0 to posEnd.
// The times parameter specifies how many times to play the loop, 0 means infinite.
//
// Returns an error if the arguments are invalid or there was a problem communicating with the device.
func (b1 *Device) PlayLoop(play bool, posStart, posEnd, times uint) error {
	// validate positions
	if err := b1.checkPatternPos(posStart); err != nil {
		return err
	}
	if err := b1.checkPatternPos(posEnd); err != nil {
		return err
	}
	// set posEnd to patt_max-1 if posEnd == 0, do this here because the firmware will do this anyway
	if posEnd == 0 {
		posEnd = getMaxPattern(b1.gen) - 1
	}

	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'p'
	buf[2] = convBoolToByte(play)
	buf[3], buf[4] = byte(posStart), byte(posEnd)
	buf[5] = byte(times & 0xff)

	// execute
	return b1.write(buf)
}

// ReadPlaystate reads the current playing state of the pattern loop.
//
// Returns the DevicePatternState struct containing the loop state and the current position,
// or an error if there was a problem communicating with the device.
func (b1 *Device) ReadPlaystate() (st DevicePatternState, err error) {
	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'S'

	// execute
	if err := b1.read(buf); err != nil {
		return st, err
	}

	// parse result
	st.IsPlaying = convByteToBool(buf[2])
	st.LoopStartPos = uint(buf[3])
	st.LoopEndPos = uint(buf[4]) // the end position from the firmware is exclusive
	st.RepeatTimes = uint(buf[5])
	st.CurrentPos = uint(buf[6])

	return st, nil
}

// SetPatternLine sets the specified pattern line to the specified DeviceLightState.
//
// The pos parameter specifies the position of the pattern line, where pos should be [0, patt_max).
// The st parameter specifies the DeviceLightState to set.
//
// Returns an error if the arguments are invalid or there was a problem communicating with the device.
func (b1 *Device) SetPatternLine(pos uint, st DeviceLightState) error {
	// validate position
	if err := b1.checkPatternPos(pos); err != nil {
		return err
	}

	// command data
	buf2 := make([]byte, cmdBufSize)
	buf2[0] = reportID
	buf2[1] = 'P'
	buf2[2], buf2[3], buf2[4] = st.R, st.G, st.B
	buf2[5], buf2[6] = convDurMsToFadeMs(st.FadeTimeMsec)
	buf2[7] = byte(pos)

	if b1.gen >= 2 {
		// set ledn for mk2+
		buf1 := make([]byte, cmdBufSize)
		buf1[0] = reportID
		buf1[1] = 'l'
		buf1[2] = st.LED.ToByte()
		// execute
		return b1.doubleWrite(buf1, buf2)
	}

	// execute for mk1
	return b1.write(buf2)
}

// ReadPatternLine reads the specified pattern line.
//
// The pos parameter specifies the position of the pattern line, where pos should be [0, patt_max).
//
// Returns the DeviceLightState struct containing the RGB values and the LEDType, or an error if there was invalid pattern position or
// a problem communicating with the device.
func (b1 *Device) ReadPatternLine(pos uint) (st DeviceLightState, err error) {
	// validate position
	if err = b1.checkPatternPos(pos); err != nil {
		return
	}

	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'R'
	buf[7] = byte(pos)

	// execute
	if err := b1.read(buf); err != nil {
		return st, err
	}

	// parse result
	st.R, st.G, st.B = buf[2], buf[3], buf[4]
	st.FadeTimeMsec = convFadeMsToDurMs(buf[5], buf[6])
	st.LED = LEDIndex(buf[7])
	return st, nil
}

// SavePattern saves the current pattern to the device.
//
// Returns an error if there was a problem communicating with the device.
func (b1 *Device) SavePattern() error {
	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'W'
	buf[2] = 0xBE
	buf[3] = 0xEF
	buf[4] = 0xCA
	buf[5] = 0xFE

	// execute and will always return error, because of issue with flash programming timing out USB
	_ = b1.write(buf)
	return nil
}

// SetTickleMode sets the device to server tickle mode, which will play the pattern from the specified start position to the specified end position after the specified timeout.
//
// The play parameter specifies whether to start or stop the tickle mode.
// The keep parameter specifies whether to keep the current pattern playing state or not.
// The posStart and posEnd parameters specify the start and end positions of the loop, where positions should be [0, patt_max-1]. If posEnd is 0, it will be set to patt_max by the firmware.
// The timeoutMsec parameter specifies the timeout in milliseconds, timeout should be at least 10ms, or it will be ignored by the firmware.
//
// Returns an error if there was a problem communicating with the device.
func (b1 *Device) SetTickleMode(play, keep bool, posStart, posEnd, timeoutMsec uint) error {
	// validate positions
	if err := b1.checkPatternPos(posStart); err != nil {
		return err
	}
	if err := b1.checkPatternPos(posEnd); err != nil {
		return err
	}
	// set posEnd to patt_max if posEnd == 0, do this here because the firmware will do the similar thing anyway
	if posEnd == 0 {
		posEnd = getMaxPattern(b1.gen)
	} else {
		// the end position from the firmware is exclusive, which is different from the play loop command. Maybe a bug?
		posEnd++
	}

	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'D'
	buf[2] = convBoolToByte(play)
	buf[3], buf[4] = convDurMsToFadeMs(timeoutMsec)
	buf[5] = convBoolToByte(keep)
	buf[6], buf[7] = byte(posStart), byte(posEnd)

	// execute
	return b1.write(buf)
}

// GetVersion returns the firmware version of the device.
//
// Returns an error if there was a problem communicating with the device.
func (b1 *Device) GetVersion() (ver int, err error) {
	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = 'v'

	// execute
	if err = b1.read(buf); err != nil {
		return
	}

	// parse result
	ver = int((buf[3]-'0')*100) + int(buf[4]-'0')
	return
}

// Test sends a test command to the device, and returns the response.
//
// Returns the response from the device, or an error if there was a problem communicating with the device.
func (b1 *Device) Test() ([]byte, error) {
	// command data
	buf := make([]byte, cmdBufSize)
	buf[0] = reportID
	buf[1] = '!'

	// execute
	err := b1.delayRead(buf, 50)
	return buf, err
}

// checkPatternPos checks if the given position is valid for the device, i.e. [0, patt_max).
// Actually, the device will not check the position value, but the arbitrary value will cause the device to play the pattern unexpectedly.
func (b1 *Device) checkPatternPos(pos uint) error {
	if maxPos := getMaxPattern(b1.gen); pos >= maxPos {
		return fmt.Errorf("b1: pattern position %d is out of range [0, %d)", pos, maxPos)
	}
	return nil
}
