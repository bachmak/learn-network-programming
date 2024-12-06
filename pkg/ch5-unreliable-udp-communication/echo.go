package ch5

import (
	"context"
	"fmt"
	"net"
)

func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
    // Create UDP socket on an address
    s, err := net.ListenPacket("udp", addr)
    if err != nil {
        return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
    }

    // Separate goroutine for echoing
    go func() {
        // Separate goroutine to wait for context cancel and stop replying
        go func() {
            <-ctx.Done()
            _ = s.Close()
        } ()

        // Prepare buffer
        buf := make([]byte, 1024)

        // Loop endlessly until error
        for {
            // Read message (from any address)
            n, clientAddr, err := s.ReadFrom(buf)
            if err != nil {
                return
            }

            // Write same message back
            _, err = s.WriteTo(buf[:n], clientAddr)
            if err != nil {
                return
            }
        }
    } ()

    return s.LocalAddr(), nil
}

