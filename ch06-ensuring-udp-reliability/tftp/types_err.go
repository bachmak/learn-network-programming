package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
)

type Err struct {
	Error   ErrCode
	Message string
}

func (e Err) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	cap := 2 + 2 + len(e.Message) + 1
	buf.Grow(cap)

	err := binary.Write(buf, binary.BigEndian, OpErr)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, e.Error)
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(e.Message)
	if err != nil {
		return nil, err
	}

	err = buf.WriteByte(0)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *Err) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)

	var opCode OpCode

	err := binary.Read(r, binary.BigEndian, &opCode)
	if err != nil {
		return err
	}

	if opCode != OpErr {
		return errors.New("invalid ERROR")
	}

	var errCode ErrCode
	err = binary.Read(r, binary.BigEndian, &errCode)
	if err != nil {
		return err
	}

	errMsg, err := r.ReadString(0)
	if err != nil {
		return err
	}

	errMsg = strings.TrimRight(errMsg, "\x00")

	e.Error = errCode
	e.Message = errMsg

	return nil
}
