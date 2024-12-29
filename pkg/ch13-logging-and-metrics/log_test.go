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

func Example_logLevels() {
	// debug logger to print to stdout with prefix DEBUG
	lDebug := log.New(
		os.Stdout,
		"DEBUG: ",
		log.Lshortfile,
	)
	// imitate log file
	logFile := new(bytes.Buffer)
	// multiwriter for the error logger (to the debug + to the log file)
	w := SustainedMultiwriter(lDebug.Writer(), logFile)
	// error logger to multiwriter with prefix ERROR
	lError := log.New(
		w,
		"ERROR: ",
		log.Lshortfile,
	)

	// print header
	fmt.Println("standard output:")
	// print to the error logger
	lError.Print("cannot communicate with the database")
	// print to the debug logger
	lDebug.Print("you cannot hum while holding your nose")

	fmt.Print("\nlog file contents:\n", logFile.String())

	// print log file contents

	// Output:
	// standard output:
	// ERROR: log_test.go:76: cannot communicate with the database
	// DEBUG: log_test.go:78: you cannot hum while holding your nose
	//
	// log file contents:
	// ERROR: log_test.go:76: cannot communicate with the database
}
