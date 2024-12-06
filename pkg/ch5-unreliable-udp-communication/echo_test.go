package ch5

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestEchoServerUDP(t *testing.T) {
    // Context to cancel echo server
    ctx, cancel := context.WithCancel(context.Background())
    // Create echo server, store address to write packets
    serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }
    // Cancel on scope exit
    defer cancel()

    // Create client socket, we'll send packets to server address
    client, err := net.ListenPacket("udp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }
    defer client.Close()

    // Send message to server
    msg := []byte("ping")
    _, err = client.WriteTo(msg, serverAddr)
    if err != nil {
        t.Fatal(err)
    }

    // Wait for server response
    buf := make([]byte, 1024)
    n, replyAddr, err := client.ReadFrom(buf)
    if err != nil {
        t.Fatal(err)
    }

    // Response has to be from the same address we sent the message to
    if replyAddr.String() != serverAddr.String() {
        t.Fatalf("received reply from %q instead of %q", replyAddr, serverAddr)
    }

    // Response should be the same as the request (echo)
    if !bytes.Equal(msg, buf[:n]) {
        t.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
    }
}
