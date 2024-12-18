package ch8

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type User struct {
	First string
	Last  string
}

// Func handler of user post requests
func handlePostUser(w http.ResponseWriter, req *http.Request) {
	// Explicitly drain the body request at scope exit
	defer func() {
		// Copy to discard and close
		_, _ = io.Copy(io.Discard, req.Body)
		_ = req.Body.Close()
	}()

	// If it's not a POST request, reply that the method is not allowed
	if req.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	// Try parsing the body into a User struct
	var u User
	err := json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		// Return bad request status if error
		msg := fmt.Sprintf("Decode failed: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Write header with accepted status
	w.WriteHeader(http.StatusAccepted)
}

// Func TestPostUser
func TestPostUser(t *testing.T) {
	// Create a test server with the only handler
	ts := httptest.NewServer(http.HandlerFunc(handlePostUser))
	// Close at scope exit
	defer func() {
		ts.Close()
	}()

	// Send a GET request to the URL
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	// Check that the method is not supported
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d; actual status %d",
			http.StatusMethodNotAllowed, resp.StatusCode,
		)
	}

	// Create a new dynamic buffer to prepare a POST request
	buf := new(bytes.Buffer)
	// Create a user and json-serialize it to the buffer
	u := User{First: "Dima", Last: "Kochetov"}
	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		t.Fatal(err)
	}

	// Make a request
	resp, err = http.Post(ts.URL, "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}
	// Check that the status is accepted
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf(
			"expected status %d; actual status %d",
			http.StatusAccepted, resp.StatusCode,
		)
	}
}
