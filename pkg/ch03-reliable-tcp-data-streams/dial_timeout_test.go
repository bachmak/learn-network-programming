package ch3

import (
	"net"
	"syscall"
	"testing"
	"time"
)

// Wrap creation of a customizable dialing object and make the dial call
func DialTimeout(
	network, address string,
	timeout time.Duration,
) (net.Conn, error) {
	// Create the dialing object
	d := net.Dialer{
		// Override control function to mock a timeout error
		Control: func(_, addr string, _ syscall.RawConn) error {
			// Always return the same error
			return &net.DNSError{
				Err:         "connection timed out",
				Name:        addr,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
		// Specify the timeout for connection establishment
		Timeout: timeout,
	}

	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	connection, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
	// Expect to see a non-nil error, since we mocked it
	if err == nil {
		connection.Close()
		t.Fatal("connection did not time out")
	}

	// "ok comma" type assert (implementation is DNSError, so net.Error)
	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}

	if !nErr.Timeout() {
		t.Fatal("error is not a timeout")
	}
}
