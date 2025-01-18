package ch4

import (
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType

	MaxPayloadSize = 10 << 20 // 10 MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload interface {
	Bytes() []byte
	fmt.Stringer
	io.WriterTo
	io.ReaderFrom
}
