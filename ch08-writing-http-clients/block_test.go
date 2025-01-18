package ch8

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func blockIndefinitely(_ http.ResponseWriter, _ *http.Request) {
	select {}
}

func TestBlockIndefinitely(t *testing.T) {
	// Would ruin the whole test if uncommented
	// ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	// _, _ = http.Get(ts.URL)
	// t.Fatal("client did not indefinitely block")
}

func TestBlockIndefinitelyWithTimeout(t *testing.T) {
	// Create a new server with an infinitely blocking handler
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))

	// Create a context with timeout of 5 second
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// Cancel at scope exit
	defer func() {
		cancel()
	}()

	// Create a new HTTP GET request specifying the context with timeout, the method, and the server's URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Execute the request using the default client
	resp, err := http.DefaultClient.Do(req)
	// Check that the error is an exceeded deadline
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return
	}
	// Close the body
	_ = resp.Body.Close()
}
