package ch3

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	// Create a cancellable context and a cancel function
	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel to sync between routines
	sync := make(chan struct{})

	// Separate routine for connection session
	go func() {
		// Notify sync on routine exit
		defer func() { sync <- struct{}{} }()

		// Create an empty dialer object
		var d net.Dialer

		// Override dialer's control method
		d.Control = func(_, _ string, _ syscall.RawConn) error {
			// Sleep for a sec
			time.Sleep(time.Second)
			// No error
			return nil
		}

		// Make connection attempt using the context
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:80")
		// There should be an error as we cancel from the main routine
		if err != nil {
			// Log it and return
			t.Log(err)
			return
		}

		// No error -> connection is active, should close it
		conn.Close()

		// But we expected an error to be present, so test didn't pass
		t.Error("connection did not time out")
	}()

	// Cancel the dial call
	cancel()

	// Wait for the routine to end
	<-sync

	// Sanity check that context was actually cancelled
	if ctx.Err() != context.Canceled {
		t.Errorf("expected cancelled context; actual: %q", ctx.Err())
	}
}
