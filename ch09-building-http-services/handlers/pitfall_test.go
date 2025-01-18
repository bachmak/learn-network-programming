package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWriteHeader(t *testing.T) {
	// create a handler that writes the body first and then the header
	handler := func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
	}

	// make an empty GET request
	req := httptest.NewRequest(http.MethodGet, "http://test/", nil)
	// create a recording response writer to inspect the response later
	rec := httptest.NewRecorder()

	// call the handler
	handler(rec, req)
	// log response status
	t.Logf("Response status: %q", rec.Result().Status)

	// create a new handler that writes the header first
	handler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
	}

	// new empty request
	req = httptest.NewRequest(http.MethodGet, "http://test/", nil)
	// new empty recorder
	rec = httptest.NewRecorder()

	// call handler, log status
	handler(rec, req)
	t.Logf("Response status: %q", rec.Result().Status)
}
