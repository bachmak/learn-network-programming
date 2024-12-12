package ch5

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func TestDialUDP(t *testing.T) {
    // Create cancellable echo server running in a separate goroutine
    ctx, cancel := context.WithCancel(context.Background())
    serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }
    defer cancel()

    // Create our UDP-client, but using net.Conn interface that filters
    // packets received from the target address
    client, err := net.Dial("udp", serverAddr.String())
    if err != nil {
        t.Fatal(err)
    }
    defer client.Close()

    // Create another UDP node, but using pure UDP packet interface
    interloper, err := net.ListenPacket("udp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    // Write a message to the initial client to check that it will
    // ignore it, since it should only receive messages from the server
    interrupt := []byte("pardon me!")
    n, err := interloper.WriteTo(interrupt, client.LocalAddr())
    if err != nil {
        t.Fatal(err)
    }
    interloper.Close()

    if n != len(interrupt) {
        t.Fatalf("wrote %d bytes of %d", n, len(interrupt))
    }

    // UDP-client to server communication
    ping := []byte("ping")
    _, err = client.Write(ping)
    if err != nil {
        t.Fatal(err)
    }

    buf := make([]byte, 1024)
    n, err = client.Read(buf)
    if err != nil {
        t.Fatal(err)
    }

    if !bytes.Equal(buf[:n], ping) {
        t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
    }

    // Wait for some additional time to make sure the client won't
    // receive any additional messages
    err = client.SetDeadline(time.Now().Add(time.Second))
    if err != nil {
        t.Fatal(err)
    }

    _, err = client.Read(buf)
    if err == nil {
        t.Fatal("unexpected packet")
    }
}
