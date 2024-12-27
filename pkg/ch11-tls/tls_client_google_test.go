package ch11

import (
	"crypto/tls"
	"net"
	"testing"
	"time"
)

// func TestClientTLSGoogle
func TestClientTLSGoogle(t *testing.T) {
	// create a new dialer with timeout of 30 second
	dialer := net.Dialer{
		Timeout: 30 * time.Second,
	}

	// create a custom TLS config
	tlsConfig := tls.Config{
		// specify preferred elliptic curve (P-256)
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
		},
		// restrict minimum TLS version to TLS 1.2
		MinVersion: tls.VersionTLS12,
	}

	// initiate a new secured connection
	// to google.com to port 443 (HTTPS default port)
	conn, err := tls.DialWithDialer(
		&dialer,
		"tcp",
		"google.com:443",
		&tlsConfig,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// print connection state: TLS version, cipher suite, certificate CA
	state := conn.ConnectionState()
	t.Logf("TLS 1.%d", state.Version-tls.VersionTLS10)
	t.Log(tls.CipherSuiteName(state.CipherSuite))
	t.Log(state.VerifiedChains[0][0].Issuer.Organization[0])
}
