package echo

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoServerUnix(t *testing.T) {
	// create a temporary dir right in the default temp directory,
	// with the name like echo_unix234if892349yds9f
	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		t.Fatal(err)
	}

	// remove the temp directory and all of its content at scope exit
	defer func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	// stop server on scope exit
	defer func() {
		cancel()
	}()

	// create filename to run echo server on; it should look like: tmp/echo_unix12345/<pid>.sock
	suffix := fmt.Sprintf("%d.sock", os.Getpid())
	socket := filepath.Join(dir, suffix)
	// start echo server
	addr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}

	// allow read/write to created socket to all users
	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	err = runClient(addr)
	if err != nil {
		t.Fatal(err)
	}
}

// func runClient
func runClient(addr net.Addr) error {
	// create unix stream client given for the socket
	client, err := net.Dial("unix", addr.String())
	if err != nil {
		return err
	}

	// close client on exit
	defer func() {
		_ = client.Close()
	}()

	// write some messages
	const msgCount = 3
	outgoing := []byte("ping")
	err = writeNTimes(client, outgoing, msgCount)
	if err != nil {
		return err
	}

	// read response
	incoming, err := read(client)
	if err != nil {
		return err
	}

	// check that response echoes the messages
	expected := bytes.Repeat(outgoing, msgCount)
	actual := incoming
	if !bytes.Equal(expected, actual) {
		return fmt.Errorf("expected reply %q; actual reply %q", expected, actual)
	}

	return nil
}

// func writeNTimes
func writeNTimes(conn net.Conn, msg []byte, n uint32) error {
	if n == 0 {
		return nil
	}

	_, err := conn.Write(msg)
	if err != nil {
		return err
	}

	return writeNTimes(conn, msg, n-1)
}

// func read
func read(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}
