package ch8

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func blockIndefinitely(_ http.ResponseWriter, _ *http.Request) {
	select {}
}

func TestBlockIndefinitely(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	_, _ = http.Get(ts.URL)
	t.Fatal("client did not indefinitely block")
}
