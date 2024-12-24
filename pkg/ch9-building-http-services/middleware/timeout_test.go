package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Func TestTimeoutMiddleware
func TestTimeoutMiddleware(t *testing.T) {
	// create a handler using timeout handler middleware, restricting handling time to one second
	handler := http.TimeoutHandler(
		http.HandlerFunc(
			// in the actual handler, write no-content header and go to sleep for a minute
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				time.Sleep(time.Minute)
			},
		),
		time.Second,
		"Timed out while reading response",
	)

	// create new empty GET request and a recorder
	req, err := http.NewRequest(http.MethodGet, "http://test/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	// call the handler
	handler.ServeHTTP(rec, req)
	// get the response and drain it at scope exit
	resp := rec.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	// check that the request timed out (service unavailable)
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status code: %q", resp.Status)
	}

	// read the body and check it matches the message from the timeout middleware
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if actual := string(b); actual != "Timed out while reading response" {
		t.Logf("unexpected body: %q", actual)
	}
}
