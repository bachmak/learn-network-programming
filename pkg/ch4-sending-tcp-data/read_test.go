package ch4

import (
	"crypto/rand"
	"testing"
    "io"
    "net"
)

// Unexported constant for server and client capacities
const serverCapacity = 1 << 24 // 16 MB
const clientCapacity = 1 << 19 // 512 KB

func TestReadIntoBuffer(t *testing.T) {
    // Create a buffer for messages sent to clients
    payload := make([]byte, serverCapacity)
    // Generate random payload data
    _, err := rand.Read(payload)
    // Fatal if error
    if err != nil {
        t.Fatal(err)
    }

    // Create listener
    listener, err := net.Listen("tcp", "127.0.0.1:")
    // Fatal if error
    if err != nil {
        t.Fatal(err)
    }

    // Spin up the listener
    go func() {
        // Accept one connection
        conn, err := listener.Accept()
        // Log and return if error
        if err != nil {
            t.Log(err)
            return
        }
        // Close connection at scope exit
        defer conn.Close()

        // Write the whole payload to the connection
        _, err = conn.Write(payload)
        // Report error if any
        if err != nil {
            t.Error(err)
        }
    } ()

    // Connect to the listener
    conn, err := net.Dial("tcp", listener.Addr().String())
    // Fatal if error
    if err != nil {
        t.Fatal(err)
    }
    // Close connection at scope exit
    defer conn.Close()

    // Create a small buffer to read from the connection
    buf := make([]byte, clientCapacity)

    // Read in a loop until EOF or error, count reads
    readCount := 0
    for {
        // Read to the buffer
        n, err := conn.Read(buf)
        // Break if error, report if not EOF
        if err != nil {
            if err != io.EOF {
                t.Error(err)
            }
            break
        }
        // Increment read counter
        readCount++

        // Log how many bytes was read
        t.Logf("%d: read %d bytes", readCount, n)
    }
}
