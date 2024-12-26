package ch4

import (
	"io"
	"net"
	"sync"
	"testing"
)

func TestProxy(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()

	server, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer conn.Close()

				// Client session
				for {
					buf := make([]byte, 1024)
					n, err := conn.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}

					switch msg := string(buf[:n]); msg {
					case "ping":
						_, err = conn.Write([]byte("pong"))
					default:
						_, err = conn.Write(buf[:n])
					}

					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
				}
			}(conn)
		}
	}()

	proxyServer, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer proxyServer.Close()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}

			go func(from net.Conn) {
				defer from.Close()

				to, err := net.Dial("tcp", server.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}
				defer to.Close()

				err = proxy(from, to)
				if err != nil && err != io.EOF {
					t.Error(err)
				}
			}(conn)
		}
	}()

	client, err := net.Dial("tcp", proxyServer.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msgs := []struct{ Message, ExpectedReply string }{
		{"ping", "pong"},
		{"pong", "pong"},
		{"echo", "echo"},
		{"pong", "pong"},
	}

	for i, msg := range msgs {
		_, err := client.Write([]byte(msg.Message))
		if err != nil {
			t.Fatal(err)
		}

		var buf [1024]byte
		n, err := client.Read(buf[:])
		if err != nil {
			t.Fatal(err)
		}

		actualReply := string(buf[:n])
		t.Logf("%q -> proxy -> %q", msg.Message, actualReply)

		if actualReply != msg.ExpectedReply {
			t.Errorf(
				"%d: expected reply: %q; actual: %q",
				i,
				msg.ExpectedReply,
				actualReply,
			)
		}
	}
}
