package main

import (
	"crypto/sha512"
	"flag"
	"fmt"
	"log"
	"os"
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s file...\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	calcAndPrintChecksum(flag.Args())
}

func calcAndPrintChecksum(args []string) {
	if len(args) == 0 {
		return
	}

	file := args[0]
	sha, err := checksum(file)
	if err != nil {
		log.Printf("[%s] checksum failed: %v", file, err)
	} else {
		fmt.Printf("[%s] checksum: %s\n", file, sha)
	}

	calcAndPrintChecksum(args[1:])
}

func checksum(file string) (string, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha512.Sum512_256(b)), nil
}
