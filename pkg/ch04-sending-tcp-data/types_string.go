package ch4

import (
	"encoding/binary"
	"io"
)

type String string

func NewString() *String { return new(String) }

func (m String) Bytes() []byte { return []byte(m) }

func (m String) String() string { return string(m) }

func (m String) WriteTo(w io.Writer) (int64, error) {
	var n int64 = 0

	err := binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}
	n += 4

	o, err := w.Write([]byte(m))
	if err != nil {
		return n, err
	}
	n += int64(o)

	return n, nil
}

func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var n int64 = 0
	var size uint32

	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return n, err
	}
	n += 4

	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	buf := make([]byte, size)
	o, err := r.Read(buf)
	if err != nil {
		return n, err
	}
	n += int64(o)
	*m = String(buf)

	return n, nil
}
