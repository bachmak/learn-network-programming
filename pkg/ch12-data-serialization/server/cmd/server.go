package main

import (
	"crypto/tls"
	"flag"
	hwproto "learn-network-programming/housework/v1"
	robotmaid "learn-network-programming/server"
	"log"
	"net"

	"google.golang.org/grpc"
)

// CLI parameters (address, certificate file, private key file)
var addr, certFile, keyFile string

// bind CLI parameters
func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
	flag.StringVar(&certFile, "cert", "cert.pem", "certificate file")
	flag.StringVar(&keyFile, "key", "key.pem", "private key file")
}

// func main
func main() {
	// parse CLI arguments
	flag.Parse()

	// create grpc server
	server := grpc.NewServer()
	// create robotmaid service instance
	rosie := new(robotmaid.Rosie)
	// register service to server
	hwproto.RegisterRobotMaidServer(server, rosie)

	// load certificate + key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// create TLS config
	tlsConfig := tls.Config{
		Certificates:             []tls.Certificate{cert},
		CurvePreferences:         []tls.CurveID{tls.CurveP256},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}

	// create TCP listener
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	// log listening
	log.Printf("Listening for TLS connections on %s ...\n", addr)

	// create TLS listener
	tlsListener := tls.NewListener(tcpListener, &tlsConfig)

	// run server
	err = server.Serve(tlsListener)
	if err != nil {
		log.Fatal(err)
	}

	// log server stopped
	log.Println("Server stopped")
}
