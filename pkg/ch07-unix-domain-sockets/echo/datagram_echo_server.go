package echo

import (
	"context"
	"net"
)

func datagramEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	// create packet-oriented listener with the given network and address
	server, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	// run asynchronously
	go func() {
		// wait for the cancel signal and stop the server
		go func() {
			<-ctx.Done()
			_ = server.Close()
		}()

		// run server
		datagramEchoServerRun(server)
	}()

	return server.LocalAddr(), nil
}

func datagramEchoServerRun(server net.PacketConn) {
	buf := make([]byte, 1024)

	datagramTalkEcho(server, buf)
}

func datagramTalkEcho(server net.PacketConn, buf []byte) {
	// read a message and echo
	n, clientAddr, err := server.ReadFrom(buf)
	if err != nil {
		return
	}

	_, err = server.WriteTo(buf[:n], clientAddr)
	if err != nil {
		return
	}

	// keep talking
	datagramTalkEcho(server, buf)
}
