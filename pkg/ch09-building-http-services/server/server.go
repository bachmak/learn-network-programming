package main

import (
	"context"
	"flag"
	"learn-network-programming/pkg/ch09-building-http-services/handlers"
	"learn-network-programming/pkg/ch09-building-http-services/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

// CLI options block:
// - listen addres,
// - dir to serve,
// - certificate,
// - private key
var (
	addr  = flag.String("listen", "127.0.0.1:8080", "listen address")
	files = flag.String("files", "./files", "static file directory")
	cert  = flag.String("cert", "", "certificate")
	pkey  = flag.String("pkey", "", "private key")
)

// func main
func main() {
	// parse CLI options
	flag.Parse()

	// run the application and log if error
	err := run(*addr, *files, *cert, *pkey)
	if err != nil {
		log.Fatal(err)
	}

	// report normal termination
	log.Println("Server gracefully shutdown")
}

// func run
func run(addr, files, cert, pkey string) error {
	// create the server, specify idle timeout and read header timeout
	server := http.Server{
		Addr:              addr,
		Handler:           buildHandler(files),
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	// channel to synchronize with the background signal listener
	done := make(chan struct{})
	// run background signal listener
	go func() {
		listenToInterrupt(&server, done)
	}()

	// report server start
	log.Printf("Serving files in %q over %s\n", files, server.Addr)

	// helper function to filter relevant errors only
	filterErrors := func(err error) error {
		if err == http.ErrServerClosed {
			return nil
		}

		return err
	}

	// if certificate and private key are specified, serve TLS
	if cert != "" && pkey != "" {
		log.Println("Serve TLS")
		return filterErrors(server.ListenAndServeTLS(cert, pkey))
	}

	return filterErrors(server.ListenAndServe())
}

// func buildHandler
func buildHandler(files string) http.Handler {
	// create a new multiplexer
	mux := http.NewServeMux()

	// serve non-hidden content from "files" directory at /static/
	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			middleware.RestrictPrefix(
				".",
				http.FileServer(http.Dir(files)),
			),
		),
	)

	// serve index.html at "/" and push the additional resources if possible
	mux.Handle(
		"/",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if pusher, ok := w.(http.Pusher); ok {
						targets := []string{
							"/static/style.css",
							"/static/hiking.svg",
						}

						for _, target := range targets {
							err := pusher.Push(target, nil)
							if err != nil {
								log.Printf("%s push failed: %v", target, err)
							}
						}
					}

					http.ServeFile(w, r, filepath.Join(files, "index.html"))
				},
			),
		},
	)

	// serve index.html without pushes at "/2"
	mux.Handle(
		"/2",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, filepath.Join(files, "index.html"))
				},
			),
		},
	)

	return mux
}

// func listenToInterrupt
func listenToInterrupt(server *http.Server, done chan struct{}) {
	// create a channel with capacity 1 (not to block notifier)
	c := make(chan os.Signal, 1)
	// register the channel to redirect SIG_INT
	signal.Notify(c, os.Interrupt)

	for {
		// wait for os.Interrupt
		if <-c == os.Interrupt {
			// gracefully shutdown the server and check errors
			if err := server.Shutdown(context.Background()); err != nil {
				log.Printf("shutdown: %v", err)
			}

			// close the sync channel and exit
			close(done)
			return
		}
	}
}
