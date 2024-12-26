package ch3

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// Create a listener on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	// Close listener automatically on the scope exit
	defer func() { _ = listener.Close() }()

	// Create a channel to wait for the goroutine to finish
	done := make(chan struct{})

	// Server
	go func() {
		for {
			// Whatever the case, prevent a deadlock
			defer func() { done <- struct{}{} }()

			// Create lisntener
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			// Client session (connection handler)
			go func(c net.Conn) {
				defer func() {
					_ = c.Close()
					done <- struct{}{}
				}()

				// Read from the connection til EOF
				buffer := make([]byte, 1024)
				for {
					n, err := c.Read(buffer)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}

					t.Logf("received: %q", buffer[:n])
				}
			}(conn)
		}
	}()

	// Client
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	// Close the connection from the client side
	conn.Close()

	// Wait for the server-side client session to finish
	<-done

	// Clise the listener
	listener.Close()

	// Wait for the listener loop to finish
	<-done
}
