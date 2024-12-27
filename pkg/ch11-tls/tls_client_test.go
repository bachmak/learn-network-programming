package ch11

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// func TestClientTLS
func TestClientTLS(t *testing.T) {
	// create a server using TLS
	server := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// check that the request is over TLS
				if r.TLS == nil {
					// redirect to the same URI, but over https
					url := "https://" + r.Host + r.RequestURI
					http.Redirect(w, r, url, http.StatusMovedPermanently)
					return
				}

				// respond with OK
				w.WriteHeader(http.StatusOK)
			},
		),
	)

	// close server at scope exit
	defer func() {
		server.Close()
	}()

	// create a client corresponding to the test server
	client := server.Client()
	// make a GET request
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	// close body at scope exit
	defer func() {
		_ = resp.Body.Close()
	}()
	// check that the status is OK
	if resp.StatusCode != http.StatusOK {
		t.Errorf(
			"expected status %d; actual %d",
			http.StatusOK,
			resp.StatusCode,
		)
	}
}
