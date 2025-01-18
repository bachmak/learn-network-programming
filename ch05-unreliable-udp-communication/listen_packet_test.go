package ch5

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestListenPacketUDP(t *testing.T) {
	// Create an echo server running in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	// Main client
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Another node to interfere client-server communication
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer interloper.Close()

	// Send interrupt message to client, though it wants to receive
	// messages from server
	interrupt := []byte("pardon me!")
	_, err = interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}

	// Client sends message to an echo server
	ping := []byte("ping")
	_, err = client.WriteTo(ping, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	// Client waits for a reply and reads one
	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the first message isn't from the server
	if addr.String() != interloper.LocalAddr().String() {
		t.Errorf(
			"expected message from %q; actual sender is %q",
			interloper.LocalAddr().String(),
			addr,
		)
	}

	// Check that it's the interrupt message
	if !bytes.Equal(interrupt, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", interrupt, buf[:n])
	}

	// Wait for a server message
	n, addr, err = client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Check that it's a server message
	if addr.String() != serverAddr.String() {
		t.Errorf(
			"expected message from %q; actual sender is %q",
			serverAddr,
			addr,
		)
	}

	// Check that it's an echo response
	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
	}
}
