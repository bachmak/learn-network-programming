package main

import (
	"flag"
	"learn-network-programming/ch06-ensuring-udp-reliability/tftp"
	"log"
	"os"
)

var (
	address = flag.String("a", "127.0.0.1:69", "listen address")
	payload = flag.String("p", "payload.svg", "file to serve to clients")
)

func main() {
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Start. Working directory: %s", dir)

	p, err := os.ReadFile(*payload)
	if err != nil {
		log.Fatal(err)
	}

	s := tftp.Server{Payload: p}
	err = s.Run(*address)
	if err != nil {
		log.Printf("Server finished with error: %v", err)
	}
}
