package ch4

import (
	"encoding/binary"
	"io"
    "errors"
)

type String string

func (m String) Bytes() []byte { return []byte(m) }

func (m String) String() string { return m }

func (m String) WriteTo(w io.Writer) (int64, error) {
    var n int64 = 0
    err := binary.Write(w, binary.BigEndian, uint8(StringType))
    if err != nil {
        return n, err
    }
    n += 1

    err = binary.Write(w, binary.BigEndian, uint32(len(m)))
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

func (m* String) ReadFrom(r io.Reader) (int64, error) {
    var n int64 = 0
    var typ uint8

    err := binary.Read(r, binary.BigEndian, &typ)
    if err != nil {
        return n, err
    }
    n += 1

    if typ != StringType {
        return n, errors.New("invalid String")
    }

    var size uint32
    err = binary.Read(r, binary.BigEndian, &size)
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
