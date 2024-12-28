package ch3

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

func Ping(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	// Initial select block (check context and an interval value from
	// the channel)
	var interval time.Duration
	select {
	case <-ctx.Done():
		return
	case interval = <-reset:
	default:
	}

	// Filter negative interval value
	if interval <= 0 {
		interval = defaultPingInterval
	}

	// Create a timer (+ prevent leaking)
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	// Wait in a loop for one of the following:
	// - context is canceled
	// - signal to reset the timer
	// - timer expires
	for {
		select {
		case <-ctx.Done():
			return
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}

			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C:
			if _, err := w.Write([]byte("ping")); err != nil {
				return
			}
		}

		_ = timer.Reset(interval)
	}
}
