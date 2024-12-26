package ch4

import (
	"io"
	"log"
	"net"
	"os"
)

type Monitor struct {
	*log.Logger
}

// Implements the io.Writer interface
func (m *Monitor) Write(p []byte) (int, error) {
	err := m.Output(2, string(p))
	// We try to avoid returning any errors not to ruin network
	// communication process
	if err != nil {
		log.Println(err)
	}

	return len(p), nil
}

func ExampleMonitor() {
	// Create a monitor printing with prefix to stdout
	monitor := &Monitor{Logger: log.New(os.Stdout, "monitor: ", 0)}

	// Synchronize with server's goroutine
	done := make(chan struct{})
	defer func() { <-done }()

	// Create listener
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		monitor.Fatal(err)
	}
	defer listener.Close()

	go func() {
		// Signal end of server work
		defer close(done)

		// Accept one connection
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Create two-way reader: to the buffer and to the passed
		// reader - monitor
		buf := make([]byte, 1024)
		r := io.TeeReader(conn, monitor)

		// Read will fill in the buffer and duplicate the data stream
		// to the monitor
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

		// Create two-way writer: to the connection and to the monitor
		w := io.MultiWriter(conn, monitor)
		// Write to the connection and to the monitor
		_, err = w.Write(buf[:n])
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}
	}()

	// Create a client
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}
	defer conn.Close()

	// Write a test message -> we should see it twice in the output:
	// once when reading from the client,
	// and then when writing the response
	_, err = conn.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}

	// Output:
	// monitor: Test
	// monitor: Test
}
