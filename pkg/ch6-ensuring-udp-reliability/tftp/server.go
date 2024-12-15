package tftp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
	Payload []byte
	Retries uint8
	Timeout time.Duration
}

func (s Server) Run(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("Listening on %s ...\n", conn.LocalAddr())

	payload := s.Payload
	retries := uint8(10)
	timeout := 6 * time.Second

	if payload == nil {
		return errors.New("payload is required")
	}

	return Server{Payload: payload, Retries: retries, Timeout: timeout}.listen(conn)
}

func (s Server) listen(conn net.PacketConn) error {
	buf := make([]byte, DatagramSize)

	_, addr, err := conn.ReadFrom(buf)
	if err != nil {
		return err
	}

	go s.runSession(addr.String(), buf)

	return s.listen(conn)
}

func (s Server) runSession(clientAddr string, msg []byte) {
	err := s.doRunSession(clientAddr, msg)
	if err != nil {
		log.Print(err)
		return
	}

	log.Printf("[%s] session ended", clientAddr)
}

func (s Server) doRunSession(clientAddr string, msg []byte) error {
	client, err := net.Dial("udp", clientAddr)
	if err != nil {
		return fmt.Errorf("[%s] dial: %v", clientAddr, err)
	}
	defer client.Close()

	state := sessionState{
		clientName: client.RemoteAddr().String(),
		block:      1,
		retries:    s.Retries,
		timeout:    s.Timeout,
	}

	rrq, err := getRequest(msg)
	if err != nil {
		log.Printf("[%s] invalid request: %v", state.clientName, err)
		err = sendError(client, ErrIllegalOp, err.Error())
		if err != nil {
			return fmt.Errorf("[%s] send error: %v", state.clientName, err)
		}
		return err
	}

	log.Printf("[%s] requested file: %s", state.clientName, rrq.Filename)

	wholeData := Data{Payload: bytes.NewReader(s.Payload)}
	return sendData(client, wholeData, state)
}

func getRequest(msg []byte) (RRQ, error) {
	var rrq RRQ
	err := rrq.UnmarshalBinary(msg)
	return rrq, err
}

func sendError(client net.Conn, errCode ErrCode, errMsg string) error {
	errPacket := Err{Error: errCode, Message: errMsg}
	data, err := errPacket.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = client.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func sendData(client net.Conn, data Data, state sessionState) error {
	blockPacket, err := data.MarshalBinary()
	if err != nil {
		if err == io.EOF {
			_, err := client.Write(blockPacket)
			if err != nil {
				return fmt.Errorf("[%s] write: %v", state.clientName, err)
			}
			return nil
		}
		return fmt.Errorf("[%s] Data.MarshalBinary: %v", state.clientName, err)
	}

	err = trySendBlock(client, blockPacket, state)
	if err != nil {
		return err
	}

	return sendData(client, data, state.incrementBlock())
}

func trySendBlock(client net.Conn, blockPacket []byte, state sessionState) error {
	if state.retries == 0 {
		return fmt.Errorf("[%s] trySendBlock: retries exhausted", state.clientName)
	}

	_, err := client.Write(blockPacket)
	if err != nil {
		return fmt.Errorf("[%s] write: %v", state.clientName, err)
	}

	buf := make([]byte, DatagramSize)
	_ = client.SetReadDeadline(time.Now().Add(state.timeout))
	_, err = client.Read(buf)

	if err == nil {
		return handleMsg(buf, state)
	}

	timeoutExpired := func() bool {
		nErr, ok := err.(net.Error)
		return ok && nErr.Timeout()
	}()

	if !timeoutExpired {
		return err
	}

	return trySendBlock(client, blockPacket, state.decrementRetries())
}
