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

func TestEchoServerUnixPacket(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unixpacket")
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

	suffix := fmt.Sprintf("%d.sock", os.Getpid())
	socket := filepath.Join(dir, suffix)
	serverAddr, err := streamingEchoServer(ctx, "unixpacket", socket)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	client, err := net.Dial("unixpacket", serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	const msgCount = 3
	msg := []byte("ping")
	err = testUnixPacketEchoServerWriteNTimes(client, msg, msgCount)
	if err != nil {
		t.Fatal(err)
	}

	err = testUnixPacketEchoServerReadNTimes(client, msg, 1024, msgCount)
	if err != nil {
		t.Fatal(err)
	}

	err = testUnixPacketEchoServerWriteNTimes(client, msg, msgCount)
	if err != nil {
		t.Fatal(err)
	}

	err = testUnixPacketEchoServerReadNTimes(client, msg[:2], 2, msgCount)
	if err != nil {
		t.Fatal(err)
	}
}

func testUnixPacketEchoServerWriteNTimes(client net.Conn, msg []byte, n uint32) error {
	if n == 0 {
		return nil
	}

	_, err := client.Write(msg)
	if err != nil {
		return err
	}

	return testUnixPacketEchoServerWriteNTimes(client, msg, n-1)
}

func testUnixPacketEchoServerReadNTimes(client net.Conn, expectedReply []byte, bufSize uint32, n uint32) error {
	if n == 0 {
		return nil
	}

	buf := make([]byte, bufSize)
	readBytes, err := client.Read(buf)
	if err != nil {
		return err
	}

	actualReply := buf[:readBytes]
	if !bytes.Equal(actualReply, expectedReply) {
		err := fmt.Errorf("expected reply %q; actual reply %q", expectedReply, actualReply)
		return err
	}

	return testUnixPacketEchoServerReadNTimes(client, expectedReply, bufSize, n-1)
}
