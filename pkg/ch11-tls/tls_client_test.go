package ch11

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/http2"
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

	testCases := []struct {
		client     *http.Client
		errMsgPart string
	}{
		{
			// client with inherent trust to the server's certificate
			server.Client(),
			"",
		},
		{
			&http.Client{
				Transport: newTransport(false, t),
			},
			"certificate signed by unknown authority",
		},
		{
			&http.Client{
				Transport: newTransport(true, t),
			},
			"",
		},
	}

	// helper function to get error messages including empty errors
	getErrorMsg := func(err error) string {
		if err == nil {
			return ""
		}
		return err.Error()
	}

	for i, testCase := range testCases {
		func() {
			// make a GET request
			resp, err := testCase.client.Get(server.URL)
			// check that error matches the expected value
			if errMsg := getErrorMsg(err); !strings.Contains(errMsg, testCase.errMsgPart) {
				t.Errorf(
					"%d: error expected to contain %q, actual %q",
					i,
					testCase.errMsgPart,
					errMsg,
				)
				return
			}
			if err != nil {
				return
			}

			// close body at scope exit
			defer func() {
				_ = resp.Body.Close()
			}()
			// check that the status is OK
			if resp.StatusCode != http.StatusOK {
				t.Errorf(
					"%d: expected status %d; actual %d",
					i,
					http.StatusOK,
					resp.StatusCode,
				)
			}
		}()
	}
}

// func newTransport
func newTransport(skipVerify bool, t *testing.T) *http.Transport {
	// create a custom HTTP transport object used for making HTTP requests
	tp := &http.Transport{
		// specify TLS config for clients
		TLSClientConfig: &tls.Config{
			// list of preferred elliptic curves used in ECDHE key exchange
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
			},
			// enable TLS 1.2 or higher only
			MinVersion: tls.VersionTLS12,
			// optionally skip certificate verification against the system's CA
			InsecureSkipVerify: skipVerify,
		},
	}

	// configure the transport to support HTTP/2, since it's not enabled
	// for custom configurations
	if err := http2.ConfigureTransport(tp); err != nil {
		t.Fatal(err)
	}

	return tp
}
