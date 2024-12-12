package ch6

// According to TFTP (RFC 1350)

const (
    HeaderSize = 4
    BlockSize = 512
    DatagramSize = HeaderSize + BlockSize
)

type OpCode uint16

const (
    OpRRQ OpCode = iota + 1
    _ // WRQ not supported
    OpData
    OpAck
    OpErr
)

type ErrCode uint16

const (
    ErrUnknown = iota
    ErrFileNotFound
    ErrAccessViolation
    ErrDiskFull
    ErrIllegalOp
    ErrUnknownID
    ErrFileExists
    ErrNoUser
)

