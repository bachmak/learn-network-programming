package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type Data struct {
	Block   uint16
	Payload io.Reader
}

func (d *Data) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Grow(DatagramSize)

	d.Block++

	err := binary.Write(buf, binary.BigEndian, OpData)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, d.Block)
	if err != nil {
		return nil, err
	}

	_, err = io.CopyN(buf, d.Payload, BlockSize)
	if err != nil {
		if err == io.EOF {
			return buf.Bytes(), err
		}
		return nil, err
	}

	return buf.Bytes(), nil
}

func (d *Data) UnmarshalBinary(p []byte) error {
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}

	var opCode OpCode

	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &opCode)
	if err != nil || opCode != OpData {
		return errors.New("invalid DATA")
	}

	err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("invalid DATA")
	}

	d.Payload = bytes.NewBuffer(p[4:])

	return nil
}
