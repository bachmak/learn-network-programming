package ch8

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
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

// Func TestMultipartPost
func TestMultipartPost(t *testing.T) {
	// Creat a new buffer for filling with request data
	buf := new(bytes.Buffer)
	// Create a multiwriter operating on the buffer
	mw := multipart.NewWriter(buf)

	// Iterate through a map with key-values for form fields
	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC1123),
		"description": "this is just my description",
	} {
		// Simply write each field to the multiwriter
		err := mw.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Then the two files
	for i, file := range []string{
		"./files/hello.txt",
		"./files/goodbye.txt",
	} {
		// First, create a file part, given a field and a file name
		fieldname := fmt.Sprintf("file%d", i)
		filename := filepath.Base(file)
		filePart, err := mw.CreateFormFile(fieldname, filename)
		if err != nil {
			t.Fatal(err)
		}

		// Open a file
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		// Copy the file's content to the filePart object
		_, err = io.Copy(filePart, f)
		// Close the file
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	// Close the writer, the request should be ready at this point
	err := mw.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Create a context with timeout (long enough, since we'll do a real request)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	// Cancel at the scope exit
	defer func() {
		cancel()
	}()

	// Create a POST request object to httpbin.org/post, with the context
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://httpbin.org/post", buf)
	if err != nil {
		t.Fatal(err)
	}
	// Set content type from the multiwriter
	req.Header.Set("Content-Type", mw.FormDataContentType())

	// Make a request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	// Drain the body at the scope exit
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read the full response to a byte-slice
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the status code is OK
	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode,
		)
	}

	// Log the response body
	t.Logf("\n%s", b)
}
