package ch3

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

const timeout = 2

func TestDialContext(t *testing.T) {
	// Init deadline 2 sec from now
	deadline := time.Now().Add(timeout * time.Second)

	// Create context which will automatically cancel after the deadline
	ctx, cancel := context.WithDeadline(context.Background(), deadline)

	// Good practice to cancel anyway to make sure the context is
	// garbage collected asap
	defer cancel()

	// Create a dialer to intercept a connect attempt and make it exceed
	// the timeout
	var dialer net.Dialer
	dialer.Control = func(_, _ string, _ syscall.RawConn) error {
		time.Sleep(timeout*time.Second + time.Millisecond)
		return nil
	}

	// Try to connect
	conn, err := dialer.DialContext(ctx, "tcp", "10.0.0.0:80")
	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}

	// Cast the error to net error and check the type
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout: %v", err)
		}
	}

	// Check the particular type of the error
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}
