package ch4

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

const payload = "The bigger the interface, the weaker the abstraction."

func TestScanner(t *testing.T) {
    // Create a listener and fatal if error
    listener, err := net.Listen("tcp", "127.0.0.1:")
    if err != nil {
        t.Fatal(err)
    }

    // Listener goroutine
    go func() {
        // Accept a connection, err and return if error
        conn, err := listener.Accept()
        if err != nil {
            t.Error(err)
            return
        }
        defer conn.Close()

        // Write the payload to the connection, err if any
        _, err = conn.Write([]byte(payload))
        if err != nil {
            t.Error(err)
        }
    } ()

    // Connect to the server, fatal if error
    conn, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()

    // Create a scanner and set a by-word splitter
    scanner := bufio.NewScanner(conn)
    scanner.Split(bufio.ScanWords)

    // Scan the words and store them in a string array
    var words []string
    for scanner.Scan() {
        words = append(words, scanner.Text())
    }

    // Check that there were no errors while scanning
    err = scanner.Err()
    if err != nil {
        t.Fatal(err)
    }

    // Expected words
    expected := []string{
        "The",
        "bigger",
        "the",
        "interface,",
        "the",
        "weaker",
        "the",
        "abstraction.",
    }

    // Compare expected and actual results
    if !reflect.DeepEqual(expected, words) {
        t.Fatalf("inaccurate scanned word list:\nexpected: %#v,\nreceived: %#v", expected, words)
    }

    // Log scanned words
    t.Logf("Scanned words: %#v", words)
}
