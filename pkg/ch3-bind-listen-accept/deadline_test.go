package ch3

import (
	"net"
	"testing"
	"time"
    "io"
)

func TestDeadline(t *testing.T) {
    // Create signal channel to sync goroutines
    sync := make(chan struct{})

    // Listener on random localhost port
    listener, err := net.Listen("tcp", "127.0.0.1:");
    // Check error, fatal if any
    if err != nil {
        t.Fatal(err)
    }

    // Separate goroutine for accepting connections
    go func() {
        // Accept connection (the only one)
        conn, err := listener.Accept()
        // Log and return if error
        if err != nil {
            t.Log(err)
            return
        }

        // At scope exit, close connection and close the channel
        // so that the other side doesn't wait
        defer func() {
            conn.Close()
            close(sync)
        }()

        // Set r/w connection's deadline (2 sec from now)
        err = conn.SetDeadline(time.Now().Add(2 * time.Second))
        // Error and return if error
        if err != nil {
            t.Error(err)
            return
        }

        // Create 1 byte capacity buffer
        buf := make([]byte, 1)
        // Read from connection to buffer (ignore byte count)
        _, err = conn.Read(buf)
        // Ok-cast error to net error
        nErr, ok := err.(net.Error)
        // Check that the error is timeout, report error otherwise
        if !ok || !nErr.Timeout() {
            t.Errorf("expected timeout error; actual: %v", err)
        }

        // Signal sync channel that we've done with reading
        sync <- struct{}{}

        // Set new deadline to enable new reads (2 sec from now)
        err = conn.SetDeadline(time.Now().Add(2 * time.Second))
        // Error and return if error
        if err != nil {
            t.Error(err)
            return
        }

        // Read from the connection to buffer -
        // now there should be some data (ignore byte count)
        _, err = conn.Read(buf)
        // If error, report it
        if err != nil {
            t.Error(err)
        }
    }()

    // Dial the listener (create connection)
    conn, err := net.Dial("tcp", listener.Addr().String())
    // Fatal if error
    if err != nil {
        t.Fatal(err)
    }
    // Close the connection at scope exit
    defer conn.Close()

    // Wait for a signal from the listener
    <-sync
    // Write 1 byte to the connection (ignore byte count)
    _, err = conn.Write([]byte("1"))
    // Fatal if error
    if err != nil {
        t.Fatal(err)
    }

    // Create an empty 1-byte buffer
    buf := make([]byte, 1)
    // Read from the connection to buffer
    _, err = conn.Read(buf)
    // Check that error is end of file (since we expect server
    // to close the connection), error if not
    if err != io.EOF {
        t.Errorf("expected server termination; actual: %v", err)
    }
}

