package ch4

import (
    "encoding/binary"
    "errors"
    "io"
)

type Binary []byte

func (m Binary) Bytes() []byte { return m }

func (m Binary) String() string { return string(m) }

func (m Binary) WriteTo(w io.Writer) (int64, error) {
    var n int64 = 0
    err := binary.Write(w, binary.BigEndian, uint8(BinaryType))
    if err != nil {
        return n, err
    }
    n += 1

    err = binary.Write(w, binary.BigEndian, uint32(len(m)))
    if err != nil {
        return n, err
    }
    n += 4

    o, err := w.Write(m)
    if err != nil {
        return n, err
    }
    n += int64(o)

    return n, nil
}

func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
    var n int64 = 0
    var typ uint8
    err := binary.Read(r, binary.BigEndian, &typ)
    if err != nil {
        return n, err
    }
    n += 1

    if typ != BinaryType {
        return n, errors.New("invalid Binary")
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

    *m = make([]byte, size)
    o, err := r.Read(*m)
    if err != nil {
        return n, err
    }
    n += int64(o)

    return n, nil
}

