package blink1

import (
	"errors"
	"testing"
	"time"
)

func TestDeviceFailedRequestDoesNotWaitForResponse(t *testing.T) {
	writeErr := errors.New("request was not sent")
	h := &transactionHID{writeErr: writeErr}
	d := &Device{dev: h}
	done := make(chan error, 1)
	go func() {
		done <- d.delayRead([]byte{1, '!'}, 60*1000, 1)
	}()
	select {
	case err := <-done:
		if !errors.Is(err, writeErr) || h.reads != 0 {
			t.Fatalf("failed request = %v, reads = %d", err, h.reads)
		}
	case <-time.After(time.Second):
		t.Fatal("failed request waited for the device response delay")
	}
}
