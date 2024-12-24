package middleware_test

import (
	"learn-network-programming/pkg/ch9-building-http-services/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Func TestRestrictPrefix
func TestRestrictPrefix(t *testing.T) {
	// create a handler which only passes through requests to /static/ endpoint
	handler := http.StripPrefix(
		"/static/",
		// use our restrict prefix middleware to restrict all hidden files
		middleware.RestrictPrefix(
			".",
			// serve files contained in the "files" subdir of the parent dir
			http.FileServer(http.Dir("../files/")),
		),
	)

	// build test cases (path to status code map)
	testCases := []struct {
		path string
		code int
	}{
		// not hidden file
		{"http://test/static/sage.svg", http.StatusOK},
		// hidden file
		{"http://test/static/.secret", http.StatusNotFound},
		// file in a hidden dir
		{"http://test/static/.dir/secret", http.StatusNotFound},
	}

	for i, testCase := range testCases {
		// create a GET request and a recorder
		req := httptest.NewRequest(http.MethodGet, testCase.path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		resp := rec.Result()

		// check the status code
		if actual := resp.StatusCode; actual != testCase.code {
			t.Errorf(
				"%d: expected %d, actual %d",
				i,
				testCase.code,
				actual,
			)
		}
	}
}
