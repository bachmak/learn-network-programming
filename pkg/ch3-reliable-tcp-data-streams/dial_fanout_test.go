package ch3

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)


func TestDialContextCancelFanOut(t *testing.T) {
    // Create a context with deadline in 10 sec after now,
    // and a cancel function
    ctx, cancel := context.WithDeadline(
        context.Background(),
        time.Now().Add(10 * time.Second),
    )

    // Create listener on localhost at random port
    listener, err := net.Listen("tcp", "127.0.0.1:")

    // If failed to create, report the error and end the test
    if err != nil {
        t.Fatal(err)
    }

    // Close listener on scope exit
    defer listener.Close()

    // Separate goroutine for accepting connections
    go func() {
        // Only accept a single connection
        conn, err := listener.Accept()

        // If okay, then close the connection,
        // we're not going to use it anyway
        if err != nil {
            conn.Close()
        }
    }()

    // Create a dial function
    dial := func(
        ctx context.Context,
        address string,
        response chan int,
        id int,
        wg *sync.WaitGroup,
    ) {
        // Signal end of work to the wait group.
        // This will decrease the counter by one
        defer wg.Done()

        // Create an empty dial object
        var d net.Dialer

        // Dial using the passed context and address
        c, err := d.DialContext(ctx, "tcp", address)

        // Only one dialer will succeed dialing; the rest will block
        // in the DialContext, until the caller call cancel on the ctx

        // If failed to dial, it's fine,
        // since the listener only accepts a single connection
        if err != nil {
            return
        }

        // We're not going to communicate,
        // so close the connection immediately
        c.Close()

        // Wait for either context being ready (expired or cancelled),
        // or another goroutine reading from the response channel

        // In other words, does the other side need the result
        // from this connection or have they already cancelled
        select {
        case <- ctx.Done():
        case response <- id:
        }
    }

    // Create a channel of ints for receiving responses from connections
    res := make(chan int)

    // Create empty wait group
    var wg sync.WaitGroup

    // Indexed loop for 10 iterations
    for i := 0; i < 10; i++ {
        // Add one to the wait group counter
        wg.Add(1)

        // Spin off a goroutine to try to dial the listener
        // and signal to the wait group
        go dial(ctx, listener.Addr().String(), res, i + 1, &wg)
    }

    // Wait for a value in the channel
    response := <-res

    // Cancel the context that was passed to the dial function
    cancel()

    // Wait for the whole wait group to finish, since we cancelled them
    wg.Wait()

    // Close the channel
    close(res)

    // Check that context was actually cancelled
    if ctx.Err() != context.Canceled {
        t.Errorf("expected canceled context; actual: %s", ctx.Err())
    }

    // Log which dialer (by index) retrieved the resource
    t.Logf("dialer %d retrieved the resource", response)
}
