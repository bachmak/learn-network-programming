package echo

import (
	"context"
	"net"
)

// main function to run streaming echo server,
// returns address on which it listens to clients
func streamingEchoServer(
	ctx context.Context,
	network string,
	addr string,
) (net.Addr, error) {
	// start listening using given network and address
	server, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	// do the rest asynchronously
	go func() {
		// asynchronously wait for cancel and then close server
		// to cancel any waiting operations on it
		go func() {
			<-ctx.Done()
			_ = server.Close()
		}()

		// start accepting connections
		acceptClients(server)
	}()

	return server.Addr(), nil
}

func acceptClients(server net.Listener) {
	// accept new client
	conn, err := server.Accept()
	if err != nil {
		return
	}

	// run session asynchronously
	go runSession(conn)

	// keep accepting new clients
	acceptClients(server)
}

func runSession(conn net.Conn) {
	// close connection at the end
	defer func() { _ = conn.Close() }()

	// start talking
	talkEcho(conn)
}

func talkEcho(conn net.Conn) {
	// read, answer with echo
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	_, err = conn.Write(buf[:n])
	if err != nil {
		return
	}

	// keep talking
	talkEcho(conn)
}
