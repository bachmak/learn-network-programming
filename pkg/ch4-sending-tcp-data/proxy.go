package ch4

import (
	"io"
	"net"
)

func proxyConn(srcAddr, dstAddr string) error {
    dstConn, err := net.Dial("tcp", dstAddr)
    if err != nil {
        return err
    }
    defer dstConn.Close()

    srcConn, err := net.Dial("tcp", srcAddr)
    if err != nil {
        return err
    }
    defer srcConn.Close()

    // srcConn <- dstConn
    go func() {
        _, _ = io.Copy(srcConn, dstConn)
        // will return when either connection is closed
    }()

    // dstConn -> srcConn
    _, err = io.Copy(dstConn, srcConn)

    return err
}

