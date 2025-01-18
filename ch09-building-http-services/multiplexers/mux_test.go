package multiplexers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// func drainAndClose (middleware handler for draining and closing
// request body after the request is served)
func drainAndClose(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// pass to the handler
			next.ServeHTTP(w, r)
			// drain and close the body
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		},
	)
}

// func TestSimpleMux
func TestSimpleMux(t *testing.T) {
	// create a new multiplexer
	serveMux := http.NewServeMux()

	// register "/" pattern, reply with no content status by default
	serveMux.HandleFunc(
		"/",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)

	// register "/hello" pattern (absolute path), reply with hello message
	serveMux.HandleFunc(
		"/hello",
		func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("Hello, friend!"))
		},
	)

	// register "/hello/there/" pattern (prefix path), reply with why hello
	serveMux.HandleFunc(
		"/hello/there/",
		func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprint(w, "Why, hello there.")
		},
	)

	// wrap the multiplexer handler in the drain and close handler
	mux := drainAndClose(serveMux)

	testCases := []struct {
		path     string
		response string
		code     int
	}{
		{
			"http://test/",
			"",
			http.StatusNoContent,
		},
		{
			"http://test/hello",
			"Hello, friend!",
			http.StatusOK,
		},
		{
			"http://test/hello/there/",
			"Why, hello there.",
			http.StatusOK,
		},
		{
			"http://test/hello/there",
			"<a href=\"/hello/there/\">Moved Permanently</a>.\n\n",
			http.StatusMovedPermanently,
		},
		{
			"http://test/hello/there/you",
			"Why, hello there.",
			http.StatusOK,
		},
		{
			"http://test/hello/and/goodbye",
			"",
			http.StatusNoContent,
		},
		{
			"http://test/something/else/entirely",
			"",
			http.StatusNoContent,
		},
		{
			"http://test/hello/you",
			"",
			http.StatusNoContent,
		},
	}

	for i, testCase := range testCases {
		func() {
			req := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)
			resp := rec.Result()
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != testCase.code {
				t.Errorf(
					"%d: expected code %d, actual %d",
					i, testCase.code, resp.StatusCode,
				)
			}

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if actual := string(b); actual != testCase.response {
				t.Errorf(
					"%d: expected response %q, actual %q",
					i, testCase.response, actual,
				)
			}
		}()
	}
}
