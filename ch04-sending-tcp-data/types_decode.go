package ch4

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func encode(w io.Writer, p Payload) (int64, error) {
	var typ uint8
	var err error
	var n int64

	switch p.(type) {
	case *Binary:
		typ = BinaryType
	case *String:
		typ = StringType
	default:
		err = errors.New("invalid payload type")
	}

	if err != nil {
		return n, err
	}

	err = binary.Write(w, binary.BigEndian, typ)
	if err != nil {
		return n, err
	}
	n += 1

	o, err := p.WriteTo(w)
	if err != nil {
		return n, err
	}
	n += o

	return n, nil
}

func decode(r io.Reader, p *Payload) (int64, error) {
	var n int64 = 0
	var typ uint8

	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return n, err
	}
	n += 1

	var payload Payload
	switch typ {
	case StringType:
		payload = NewString()
	case BinaryType:
		payload = NewBinary()
	default:
		msg := fmt.Sprintf("invalid type: %d", typ)
		err = errors.New(msg)
	}

	if err != nil {
		return n, err
	}

	o, err := payload.ReadFrom(r)
	if err != nil {
		return n, err
	}
	n += o

	*p = payload

	return n, nil
}
