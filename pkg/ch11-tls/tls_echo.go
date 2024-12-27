package ch11

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// func NewTLSServer
func NewTLSServer(
	ctx context.Context,
	addr string,
	idleTimeout time.Duration,
	tlsConfig *tls.Config,
) *Server {
	return &Server{
		ctx:         ctx,
		addr:        addr,
		idleTimeout: idleTimeout,
		tlsConfig:   tlsConfig,
	}
}

// struct Server
type Server struct {
	// context to stop the server
	ctx context.Context
	// channel to wait until server is ready to accept connections
	ready chan struct{}
	// listen address
	addr string

	// idle timeout to discard slow clients
	idleTimeout time.Duration
	tlsConfig   *tls.Config
}

// func Ready to wait until server is ready
func (s *Server) Ready() {
	if s.ready != nil {
		<-s.ready
	}
}

// func ListenAndServeTLS
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	// do tcp-related initialization
	if err := s.initTCP(); err != nil {
		return fmt.Errorf("init TCP: %v", err)
	}

	// create a TCP-listener on the specified address
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("binding to tcp %s: %w", s.addr, err)
	}

	// if server is stoppable, close it on stop asynchronously
	if s.ctx != nil {
		go func() {
			<-s.ctx.Done()
			_ = listener.Close()
		}()
	}

	// serve
	return s.ServeTLS(listener, certFile, keyFile)
}

// func ServeTLS
func (s *Server) ServeTLS(listener net.Listener, certFile, keyFile string) error {
	// do TLS-related initialization
	if err := s.initTLS(certFile, keyFile); err != nil {
		return fmt.Errorf("init TLS: %v", err)
	}

	// create a new TLS listener over the TCP one
	tlsListener := tls.NewListener(listener, s.tlsConfig)

	// accept new connections in a loop
	for {
		conn, err := tlsListener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %v", err)
		}

		// run client session asynchronously
		go s.runClientSession(conn)
	}
}

// func initTCP
func (s *Server) initTCP() error {
	// set default address if not set
	if s.addr == "" {
		s.addr = "localhost:4043"
	}

	// reset ready channel if presented
	if s.ready != nil {
		close(s.ready)
	}

	return nil
}

// func initTLS
func (s *Server) initTLS(certFile, keyFile string) error {
	// set TLS config if hasn't been
	if s.tlsConfig == nil {
		// elliptic curve P-256
		s.tlsConfig = &tls.Config{
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
			},
			// TLS minimum TLS 1.2
			MinVersion: tls.VersionTLS12,
			// prefer server-side ciphers
			PreferServerCipherSuites: true,
		}
	}

	// if certificate list is empty and there's no function to provide them
	if len(s.tlsConfig.Certificates) == 0 && s.tlsConfig.GetCertificate == nil {
		// load certificate from key pair
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("loading certificate: %v", err)
		}

		// set certificate list to the TLS config
		s.tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return nil
}

// func runClientSession
func (s *Server) runClientSession(conn net.Conn) {
	// close connection at scope exit
	defer func() {
		_ = conn.Close()
	}()

	// echo in a loop
	for {
		// if idle timeout is provided, push the read/write deadline forward
		if s.idleTimeout != 0 {
			err := conn.SetDeadline(time.Now().Add(s.idleTimeout * time.Second))
			if err != nil {
				return
			}
		}

		// read a message
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		// echo the message
		_, err = conn.Write(buf[:n])
		if err != nil {
			return
		}
	}
}
