package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"learn-network-programming/ch13-logging-and-metrics/instrumentation/metrics"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// CLI options
var (
	// address to serve metrics (for the prometheus handler)
	metricsAddr = flag.String("metrics", "127.0.0.1:8081", "metrics listen address")
	// address to serve main logic (for our custom handler)
	webAddr = flag.String("web", "127.0.0.1:8082", "web listen address")
)

// func helloHandler
func helloHandler(w http.ResponseWriter, _ *http.Request) {
	// increase metrics' request counter
	metrics.Requests.Add(1)
	// time the request and add the duration to the histogram when the processing is finished
	defer func(start time.Time) {
		duration := time.Since(start).Seconds()
		metrics.RequestDuration.Observe(duration)
	}(time.Now())

	// write hello
	_, err := w.Write([]byte("Hello!"))
	// add error count if any
	if err != nil {
		metrics.WriteErrors.Add(1)
	}
}

// func newHTTPServerAsync
func newHTTPServerAsync(
	addr string,
	handler http.Handler,
	stateFunc func(net.Conn, http.ConnState),
) error {
	// create tcp listener at the given address
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// create a server
	server := &http.Server{
		// address
		Addr: addr,
		// handler
		Handler: handler,
		// timeout to keep the connection alive
		IdleTimeout: time.Minute,
		// timeout to wait for the header
		ReadHeaderTimeout: 30 * time.Second,
		// callback function to notify connection state changes
		ConnState: stateFunc,
	}

	// run server asynchronously
	go func() {
		err := server.Serve(listener)
		log.Fatal(err)
	}()

	return nil
}

// func connStateMetrics
func connStateMetrics(_ net.Conn, state http.ConnState) {
	// since we're only interested in connection count, we ignore
	// the connection itself and simply inspect if it has been
	// opened or closed
	//
	// increase connection count for a new connection,
	// decrease for a closed one
	switch state {
	case http.StateNew:
		metrics.OpenConnections.Add(1)
	case http.StateClosed:
		metrics.OpenConnections.Add(-1)
	}
}

// func main
func main() {
	// parse CLI options
	flag.Parse()

	// server part
	//
	// create a multiplexer to serve metrics on a separate endpoint
	mux := http.NewServeMux()
	// delegate metrics serving to prometheus
	mux.Handle("/metrics/", promhttp.Handler())
	// run a separate server to serve metrics from the main service
	if err := newHTTPServerAsync(*metricsAddr, mux, nil); err != nil {
		log.Fatal(err)
	}

	// report metrics server start
	fmt.Printf("Metrics listening on %q ...\n", *metricsAddr)

	// run the main server, inject the connection metrics callback
	if err := newHTTPServerAsync(*webAddr, http.HandlerFunc(helloHandler), connStateMetrics); err != nil {
		log.Fatal(err)
	}

	// report main server start
	fmt.Printf("Web listening on %q ...\n\n", *webAddr)

	// client part
	//
	// 500 clients * 100 GET requests each,
	// wait group to wait for clients' goroutines
	clients := 500
	gets := 100
	wg := &sync.WaitGroup{}

	// report client spawning
	fmt.Printf(
		"Spawning %d clients to make %d requests each ...",
		clients,
		gets,
	)

	// spawn clients
	for i := 0; i < clients; i++ {
		// increase wait group wait counter
		wg.Add(1)
		// run client in a goroutine
		go func() {
			// decrease wait group counter when finished
			defer func() {
				wg.Done()
			}()

			// create a client with explicit cloning of the underlying
			// transport layer to prevent it from reusing the same TCP sockets
			client := &http.Client{
				Transport: http.DefaultTransport.(*http.Transport).Clone(),
			}

			// make j get requests
			for j := 0; j < gets; j++ {
				url := fmt.Sprintf("http://%s/", *webAddr)
				resp, err := client.Get(url)
				if err != nil {
					log.Fatal(err)
				}

				// explicitly read the body and close it
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}
		}()
	}

	// wait for all clients to finish
	wg.Wait()
	// report done
	fmt.Print(" done.\n\n")

	// metrics client part
	//
	// make a GET request to the metrics endpoint (of the metrics server)
	url := fmt.Sprintf("http://%s/metrics", *metricsAddr)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	// close response body at scope exit
	defer func() {
		_ = resp.Body.Close()
	}()

	// read response to a byte slice
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// build metrics prefix to filter prometheus metrics based on the namespace and subsystem
	metricsPrefix := fmt.Sprintf("%s_%s", *metrics.Namespace, *metrics.Subsystem)
	// print header
	fmt.Print("Current Metrics:\n")

	// iterate through the metrics lines
	for _, line := range bytes.Split(b, []byte("\n")) {
		// print only those that have the prefix
		if bytes.HasPrefix(line, []byte(metricsPrefix)) {
			fmt.Printf("%s\n", line)
		}
	}
}
