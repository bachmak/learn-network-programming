package ch3

import (
	"context"
	"fmt"
	"time"
    "io"
)

func ExamplePing() {
    // Cancellabel context to pass to the ping function
    ctx, cancel := context.WithCancel(context.Background())
    // Pipe instead of a real connection
    r, w := io.Pipe()
    // Channel to synchronize goroutines
    done := make(chan struct{})
    // Buffered channel (1 empty slot)
    resetTimer := make(chan time.Duration, 1)
    // Non-blocking write to channel
    resetTimer <- time.Second

    // Separate goroutine to do the pinging
    go func() {
        // Do unless cancelled
        Ping(ctx, w, resetTimer)
        // Signal that we're done
        close(done)
    }()

    // Helper function to change timeouts and read from the "connection"
    receivePing := func(d time.Duration, r io.Reader) {
        // Reset the timeout
        if d >= 0 {
            fmt.Printf("resetting timer (%s)\n", d)
            resetTimer <- d
        }

        // Time the read
        now := time.Now()
        buf := make([]byte, 1024)
        n, err := r.Read(buf)

        if err != nil {
            fmt.Println(err)
        }

        // Print the message and how much time it took to receive it
        fmt.Printf("received %q (%s)\n", buf[:n], time.Since(now).Round(100 * time.Millisecond))
    }

    // Iterate through a range of millisecond durations
    // (negative ones don't change the timeout)
    for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
        fmt.Printf("Run %d:\n", i+1)
        receivePing(time.Duration(v) * time.Millisecond, r)
    }

    // Stop the ping function
    cancel()
    // Wait for the goroutine to finish
    <-done

    // Output:
    // Run 1:
    // resetting timer (0s)
    // received "ping" (1s)
    // Run 2:
    // resetting timer (200ms)
    // received "ping" (200ms)
    // Run 3:
    // resetting timer (300ms)
    // received "ping" (300ms)
    // Run 4:
    // resetting timer (0s)
    // received "ping" (300ms)
    // Run 5:
    // received "ping" (300ms)
    // Run 6:
    // received "ping" (300ms)
    // Run 7:
    // received "ping" (300ms)
}
