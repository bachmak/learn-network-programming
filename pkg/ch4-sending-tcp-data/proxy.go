package ch4

import (
	"io"
	"net"
)

func proxy(src io.Reader, dst io.Writer) error {
    srcWriter, srcIsWriter := src.(io.Writer)
    dstReader, dstIsReader := dst.(io.Reader)

    if srcIsWriter && dstIsReader {
        go func () {
            _, _ = io.Copy(srcWriter, dstReader)
        }()
    }

    _, err := io.Copy(dst, src)

    return err
}

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

    return proxy(srcConn, dstConn)
}

