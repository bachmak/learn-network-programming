package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
)

type RRQ struct {
	Filename string
	Mode     string
}

func (q RRQ) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	cap := 2 + len(q.Filename) + 1 + len(q.Mode) + 1
	buf.Grow(cap)

	err := binary.Write(buf, binary.BigEndian, OpRRQ)
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(q.Filename)
	if err != nil {
		return nil, err
	}

	err = buf.WriteByte(0)
	if err != nil {
		return nil, err
	}

	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}

	_, err = buf.WriteString(mode)
	if err != nil {
		return nil, err
	}

	err = buf.WriteByte(0)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (q *RRQ) UnmarshalBinary(p []byte) error {
	buf := bytes.NewBuffer(p)

	var code OpCode
	err := binary.Read(buf, binary.BigEndian, &code)
	if err != nil {
		return err
	}

	if code != OpRRQ {
		return errors.New("invalid RRQ")
	}

	filename, err := buf.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	filename = strings.TrimRight(filename, "\x00")
	if len(filename) == 0 {
		return errors.New("invalid RRQ")
	}

	mode, err := buf.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	mode = strings.TrimRight(mode, "\x00")
	mode = strings.ToLower(mode)
	if mode != "octet" {
		return errors.New("only binary transfers supported")
	}

	q.Filename = filename
	q.Mode = mode

	return nil
}
