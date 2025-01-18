package ch8

import (
	"net/http"
	"testing"
	"time"
)

func TestHeadTime(t *testing.T) {
	// Make https HEAD request to time.gov
	resp, err := http.Head("https://time.gov/")
	if err != nil {
		t.Fatal(err)
	}
	// Close body response at the scope exit
	defer func() {
		_ = resp.Body.Close()
	}()

	// Obtain Date key value from the response header
	serverDateValue := resp.Header.Get("Date")
	// Report error if empty
	if serverDateValue == "" {
		t.Fatal("no Date header received from time.gov")
	}

	// Parse time according to RFC1123
	serverTime, err := time.Parse(time.RFC1123, serverDateValue)
	if err != nil {
		t.Fatal(err)
	}

	// Get current time rounded to seconds
	localTime := time.Now().Round(time.Second)
	// Show server time and the difference between the local time
	t.Logf("time.gov: %s (skew %s)", serverTime, localTime.Sub(serverTime))
}
