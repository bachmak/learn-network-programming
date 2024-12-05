package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

// Global variables - addresses of CLI parameters
var (
    // Number of tries
    count = flag.Int("c", 3, "number of pings: <= 0 means forever")
    // Interval between pings
    interval = flag.Duration("i", time.Second, "inteval between pings")
    // Connection timeout
    timeout = flag.Duration("W", 5*time.Second, "time to wait for a reply")
)

// Init function - called before main
func init() {
    // Initialize usage function (print description and default values)
    flag.Usage = func() {
        fmt.Printf("Usage: %s [options] host:port\nOptions:\n", os.Args[0])
        flag.PrintDefaults()
    }
}

func main() {
    // Parse CLI arguments
    flag.Parse()

    // If not enough arguments: print warning, show usage, and exit
    if flag.NArg() <= 1 {
        fmt.Printf("host:port is required\n\n")
        flag.Usage()
        os.Exit(1)
    }

    // String for address to ping to
    target := flag.Args()[1]
    // Show aknowledge message
    fmt.Println("PING", target)

    // If no count limit, warn that user should send interrupt to exit
    if *count <= 0 {
        fmt.Println("Press CTRL+C to exit.")
    }

    // Ping counter (0 initial value)
    msg := 0

    // Loop if either ping endlessly or count exhausted
    for (*count <= 0) || (msg < *count) {
        // Increment try count
        msg++
        // Indicate try
        fmt.Print(msg, " ")

        // Start measuring time from now to end of dial try
        start := time.Now()
        // Dial with timeout specified
        c, err := net.DialTimeout("tcp", target, *timeout)
        // Stop measuring time
        dur := time.Since(start)

        // If any error
        if err != nil {
            // Print that we've failed
            fmt.Printf("faile in %s: %v\n", dur, err)
            // Cast error to network error, exit if it's not temporary
            if nErr, ok := err.(net.Error); !ok || !nErr.Temporary() {
                os.Exit(1)
            }
        // Else close connection and print duration
        } else {
            _ = c.Close()
            fmt.Print(dur)
        }

        // Sleep for time specified in interval
        time.Sleep(*interval)
    }
}
