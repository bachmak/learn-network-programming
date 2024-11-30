package ch3

import (
	"context"
	"net"
	"testing"
    "io"
    "time"
)

func TestPingAdvanceDeadline(t *testing.T) {
    // Create a channel for synchronization
    done := make(chan struct{})
    // Start a listener that accepts a connection
    listener, err := net.Listen("tcp", "127.0.0.1:")
    // Fatal if error
    if err != nil {
        t.Fatal(err)
    }

    // Initial timestamp to time messages
    begin := time.Now()

    // Goroutine for accepting a connection
    go func() {
        // Close sync channel regardless of what happens
        defer func() {
            close(done)
        } ()

        // Accept one connection
        conn, err := listener.Accept()
        // Log and return if error
        if err != nil {
            t.Log(err)
            return
        }

        // Create cancelable context
        ctx, cancel := context.WithCancel(context.Background())
        // Cancel all operations in case of early exit
        // and close connection
        defer func() {
            cancel()
            conn.Close()
        } ()

        // Create a 1-capacity channel for resetting timer
        resetTimer := make(chan time.Duration, 1)
        // Write initial timeout of one second
        resetTimer <- time.Second
        // Spin off ping function to ping the connection
        go Ping(ctx, conn, resetTimer)

        // Set rw deadline on connection, 5 sec from now
        err = conn.SetDeadline(time.Now().Add(5 * time.Second))
        // Report error and return if any
        if err != nil {
            t.Error(err)
            return
        }

        // Create a buffer to read from connection
        buf := make([]byte, 1024)
        // Read in loop
        for {
            // Read
            n, err := conn.Read(buf)
            // Just return if error
            if err != nil {
                return
            }
            // Log time since the beginning (truncate to seconds)
            // and the buffer content
            t.Logf(
                "[%s] %s",
                time.Since(begin).Truncate(time.Second),
                buf[:n],
            )

            // Reset the ping timer (without overriding its timeout value)
            resetTimer <- 0

            // Advance connection deadline since we've read some data
            err = conn.SetDeadline(time.Now().Add(5 * time.Second))
            // Report error and return if any
            if err != nil {
                t.Error(err)
                return
            }
        }
    } ()

    // Dial the server
    conn, err := net.Dial("tcp", listener.Addr().String())
    // Fatal if error
    if err != nil {
        t.Fatal(err)
    }
    // Close connection at scope exit
    defer conn.Close()

    // Create buffer to read to (from pinger)
    buf := make([]byte, 1024)
    // Read four times
    for i := 0; i < 4; i++ {
        // Read
        n, err := conn.Read(buf)
        // Fatal if error
        if err != nil {
            t.Fatal(err)
        }
        // Log time passed since begin (in seconds) and the buffer content
        t.Logf(
            "[%s] %s",
            time.Since(begin).Truncate(time.Second),
            buf[:n],
        )
    }

    // Answer after the fourth ping message with pong
    _, err = conn.Write([]byte("PONG!!!"))
    // Fatal if error
    if err != nil {
        t.Fatal(err)
    }

    // Read another four messages until EOF
    for i := 0; i < 4; i++ {
        // Read
        n, err := conn.Read(buf)
        // Fatal if non EOF error
        if err != nil {
            if err != io.EOF {
                t.Fatal(err)
            }
        }

        // Log timestamp and the buffer content
        t.Logf(
            "[%s] %s",
            time.Since(begin).Truncate(time.Second),
            buf[:n],
        )
    }

    // Wait for the server to exit its goroutine
    // as the client doesn't answer to the ping messages anymore
    <-done
    // End timestamp to get the total duration (in sec)
    end := time.Since(begin).Truncate(time.Second)
    // Log timestamp
    t.Logf("[%s] done", end)
    // Check that it's 9 seconds
    // (4 pings + 4 pings + 1 sec of closing connection by timeout)
    if end != 9 * time.Second {
        t.Fatalf("expected EOF at 9 seconds; actual %s", end)
    }
}

