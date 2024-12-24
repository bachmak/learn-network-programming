package ch9

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"learn-network-programming/pkg/ch9-building-http-services/handlers"
	"net"
	"net/http"
	"testing"
	"time"
)

// func runSimpleHTTPServer
func runSimpleHTTPServer(addr string, done chan error, ctx context.Context) {
	// Create a server
	server := &http.Server{
		// use the given address
		Addr: addr,
		// wrap the single default handler in a timeout handler
		Handler: http.TimeoutHandler(
			// use the default methods handler
			handlers.DefaultMethodsHandler(),
			// handler execution timeout
			2*time.Minute,
			// use default message
			"",
		),
		// timeout to keep tcp connection alive
		IdleTimeout: 5 * time.Minute,
		// timeout to wait between a new client connects and sends the header
		ReadHeaderTimeout: time.Minute,
	}

	// create a new tcp listener on the server's address
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		done <- err
		return
	}

	// serve clients asynchronously
	go func() {
		// listen to context cancel and close the server
		go func() {
			<-ctx.Done()
			server.Close()
		}()

		done <- server.ServeTLS(listener, "ch9.pem", "ch9-key.pem")
	}()
}

// func TestSimpleHTTPServer
func TestSimpleHTTPServer(t *testing.T) {
	// create a channel for getting errors from the server
	done := make(chan error)
	// create a context to cancel the server
	ctx, cancel := context.WithCancel(context.Background())
	// address to run server at
	addr := "127.0.0.1:8443"
	// run the server asynchronously
	runSimpleHTTPServer(addr, done, ctx)
	// cancel, wait for an error in the channel at the exit scope
	defer func() {
		cancel()
		err := <-done
		// if server is closed, it's okay
		if err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}()

	// array slice with test cases
	testCases := []struct {
		method   string
		body     io.Reader
		code     int
		response string
	}{
		// test for an empty get request
		{
			http.MethodGet,
			nil,
			http.StatusOK,
			"Hello, friend!",
		},
		// test for a post request with an html-like body
		{
			http.MethodPost,
			bytes.NewBufferString("<world>"),
			http.StatusOK,
			"Hello, &lt;world&gt;!",
		},
		// test for an unsupported method request
		{
			http.MethodHead,
			nil,
			http.StatusMethodNotAllowed,
			"",
		},
	}

	// create a client a URL
	url := fmt.Sprintf("https://%s/", addr)
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// for each test case
	for i, testCase := range testCases {
		// additional block for proper deferring
		func() {
			// create a new HTTP request based on the test case
			request, err := http.NewRequest(testCase.method, url, testCase.body)
			if err != nil {
				t.Errorf("%d: %v", i, err)
				return
			}

			// send the request and wait for the response
			response, err := client.Do(request)
			if err != nil {
				t.Errorf("%d: %v", i, err)
				return
			}
			// drain the body at the scope exit
			defer func() {
				_ = response.Body.Close()
			}()

			// check status code
			if response.StatusCode != testCase.code {
				t.Errorf(
					"%d: expected status = %q, actual = %q",
					i, testCase.code, response.StatusCode,
				)
				return
			}

			// check content
			content, err := io.ReadAll(response.Body)
			if err != nil {
				t.Errorf("%d: %v", i, err)
				return
			}

			if string(content) != testCase.response {
				t.Errorf(
					"%d: expected response = %q, actual = %q",
					i, testCase.response, content,
				)
				return
			}
		}()
	}
}
