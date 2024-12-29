package main

import (
	"flag"
	hwproto "learn-network-programming/housework/v1"
	robotmaid "learn-network-programming/server"
	"log"
	"net"

	"google.golang.org/grpc"
)

// CLI parameters
var addr string

// bind CLI parameters
func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
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

	// create TCP listener
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	// log listening
	log.Printf("Listening for TLS connections on %s ...\n", addr)

	// run server
	err = server.Serve(tcpListener)
	if err != nil {
		log.Fatal(err)
	}

	// log server stopped
	log.Println("Server stopped")
}
