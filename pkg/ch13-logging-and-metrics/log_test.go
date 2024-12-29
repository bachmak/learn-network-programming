package ch13

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

// func Example_log
func Example_log() {
	// create logger to stdout with custom prefix and logging file name only
	l := log.New(
		os.Stdout,
		"example: ",
		log.Lshortfile,
	)
	// log something
	l.Print("logging to standard output")

	// Output:
	// example: log_test.go:19: logging to standard output
}

// func Example_logMultiWriter
func Example_logMultiWriter() {
	// imitate log file with byte buffer
	logFile := new(bytes.Buffer)
	// create sustained multiwriter to stdout and log file
	w := SustainedMultiwriter(os.Stdout, logFile)
	// create logger based on mw, with custom prefix (leftmost)
	// and logging short file name
	l := log.New(
		w,
		"example: ",
		log.Lshortfile|log.Lmsgprefix,
	)

	// print header to stdout
	fmt.Println("standard output:")
	// log something
	l.Print("Canada is south of Detroit")

	// print log file contents
	fmt.Print("\nlog file contents:\n", logFile.String())

	// Output:
	// standard output:
	// log_test.go:42: example: Canada is south of Detroit
	//
	// log file contents:
	// log_test.go:42: example: Canada is south of Detroit
}
