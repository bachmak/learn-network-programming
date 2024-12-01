package ch4

import (
	"encoding/binary"
	"errors"
	"io"
)

func decode(r io.Reader) (Payload, error) {
    var typ uint8

    err := binary.Read(r, binary.BigEndian, &typ)
    if err != nil {
        return nil, err
    }

    var payload Payload
    switch typ {
    case StringType:
        payload = NewString()
    case BinaryType:
        payload = NewBinary()
    default:
        err = errors.New("unknown type")
    }

    if err != nil {
        return nil, err
    }

    _, err = payload.ReadFrom(r)
    if err != nil {
        return nil, err
    }

    return payload, nil
}

