package ch11

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// func TestEchoServerTLS
func TestEchoServerTLS(t *testing.T) {
	// create a context to cancel the server
	ctx, cancel := context.WithCancel(context.Background())

	// separate variables to avoid reading server state as it runs in a goroutine
	addr := "localhost:34443"
	idleTimeout := 200 * time.Millisecond

	// create the server and delegate it configuring the TLS
	server := NewTLSServer(ctx, addr, idleTimeout, nil)

	// channel to correctly wait for the server goroutine to finish
	done := make(chan struct{})
	// run the server in a goroutine
	go runEchoServerTLS(server, done, t)
	// wait for its initialization to finish
	server.Ready()

	// run echo client
	err := runEchoClientTLS(addr, idleTimeout)
	if err != nil {
		t.Fatal(err)
	}

	// stop the server
	cancel()
	// wait for the server to finish
	<-done
}

// func runEchoServerTLS
func runEchoServerTLS(server *Server, done chan struct{}, t *testing.T) {
	// notify finish at scope exit
	defer func() {
		done <- struct{}{}
	}()

	// listen and serve
	err := server.ListenAndServeTLS("cert.pem", "key.pem")
	// check that server finished correctly
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		t.Error(err)
	}
}

// func runEchoClientTLS
func runEchoClientTLS(addr string, idleTimeout time.Duration) error {
	// init TLS config
	tlsConfig, err := initTLSConfig()
	if err != nil {
		return fmt.Errorf("init TLS config: %v", err)
	}

	// connect to the server using the TLS config
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("dial TLS: %v", err)
	}
	// close connection at scope exit
	defer func() {
		_ = conn.Close()
	}()

	// send hello
	hello := []byte("hello")
	_, err = conn.Write(hello)
	if err != nil {
		return fmt.Errorf("write hello: %v", err)
	}

	// read response and check that it's the same as the request
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("read echo: %v", err)
	}

	if actual := buf[:n]; !bytes.Equal(actual, hello) {
		return fmt.Errorf("expected response %q; actual %q", hello, actual)
	}

	// go to sleep for long enough so that server closes the connection
	time.Sleep(2 * idleTimeout)
	_, err = conn.Read(buf)
	if err != io.EOF {
		return fmt.Errorf("expected EOF; actual %v", err)
	}

	return nil
}

// func initTLSConfig
func initTLSConfig() (*tls.Config, error) {
	// read certificate file
	cert, err := os.ReadFile("cert.pem")
	if err != nil {
		return nil, fmt.Errorf("read cert: %v", err)
	}

	// create a new certificate pool and add our certificate to it
	certPool := x509.NewCertPool()

	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		return nil, fmt.Errorf("failed to add cert")
	}

	// create a new TLS config,
	// specify elliptic curvese, TLS version, and the certificate pool
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion:       tls.VersionTLS12,
		RootCAs:          certPool,
	}

	return tlsConfig, nil
}
