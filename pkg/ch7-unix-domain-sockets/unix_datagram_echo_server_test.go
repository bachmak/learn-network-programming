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

type testUnixgramEcho struct{}

func TestDatagramEchoServer(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unixgram")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Error(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	serverAddr, err := testUnixgramEcho{}.createAndRunServerAsync(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}

	err = testUnixgramEcho{}.createAndRunClient(serverAddr, dir)
	if err != nil {
		t.Fatal(err)
	}
}

func (testUnixgramEcho) createAndRunServerAsync(ctx context.Context, dir string) (net.Addr, error) {
	suffix := fmt.Sprintf("server%d.sock", os.Getpid())
	socket := filepath.Join(dir, suffix)
	addr, err := datagramEchoServer(ctx, "unixgram", socket)
	if err != nil {
		return nil, err
	}

	err = os.Chmod(socket, os.ModeSocket|0622)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func (testUnixgramEcho) createAndRunClient(serverAddr net.Addr, dir string) error {
	suffix := fmt.Sprintf("client%d.sock", os.Getpid())
	socket := filepath.Join(dir, suffix)

	client, err := net.ListenPacket("unixgram", socket)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	err = os.Chmod(socket, os.ModeSocket|0622)
	if err != nil {
		return err
	}

	err = testUnixgramEcho{}.runClient(client, serverAddr)
	if err != nil {
		return err
	}

	return nil
}

func (testUnixgramEcho) runClient(client net.PacketConn, serverAddr net.Addr) error {
	const msgCount = 3
	outgoing := []byte("ping")
	err := testUnixgramEcho{}.writeNTimes(client, serverAddr, outgoing, msgCount)
	if err != nil {
		return err
	}

	readContext := testUnixgramEchoReadContext{
		client:        client,
		expectedAddr:  serverAddr,
		expectedReply: outgoing,
	}
	err = readContext.readNTimes(msgCount)
	if err != nil {
		return err
	}

	return nil
}

func (testUnixgramEcho) writeNTimes(client net.PacketConn, addr net.Addr, msg []byte, n uint32) error {
	if n == 0 {
		return nil
	}

	_, err := client.WriteTo(msg, addr)
	if err != nil {
		return err
	}

	return testUnixgramEcho{}.writeNTimes(client, addr, msg, n-1)
}

type testUnixgramEchoReadContext struct {
	client        net.PacketConn
	expectedAddr  net.Addr
	expectedReply []byte
}

func (c testUnixgramEchoReadContext) readNTimes(n uint32) error {
	if n == 0 {
		return nil
	}

	buf := make([]byte, 256)
	nBytes, actualAddr, err := c.client.ReadFrom(buf)
	if err != nil {
		return err
	}

	if actualAddr.String() != c.expectedAddr.String() {
		err := fmt.Errorf("received reply from %q instead of %q", actualAddr, c.expectedAddr)
		return err
	}

	actualReply := buf[:nBytes]
	if !bytes.Equal(actualReply, c.expectedReply) {
		err := fmt.Errorf("expected reply %q; actual reply %q", c.expectedReply, actualReply)
		return err
	}

	return c.readNTimes(n - 1)
}
