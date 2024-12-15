package ch6

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type Ack uint16

func (a Ack) MarshalBinary() ([]byte, error) {
    buf := new(bytes.Buffer)
    cap := 2 + 2
    buf.Grow(cap)

    err := binary.Write(buf, binary.BigEndian, OpAck)
    if err != nil {
        return nil, err
    }

    err = binary.Write(buf, binary.BigEndian, a)
    if err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}

func (a *Ack) UnmarshalBinary(p []byte) error {
    r := bytes.NewReader(p)

    var opCode OpCode
    err := binary.Read(r, binary.BigEndian, &opCode)
    if err != nil {
        return err
    }

    if opCode != OpAck {
        return errors.New("invalid ACK")
    }

    var block uint16
    err = binary.Read(r, binary.BigEndian, &block)
    if err != nil {
        return err
    }

    *a = Ack(block)
    return nil
}

