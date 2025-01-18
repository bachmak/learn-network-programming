package ch11

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// func TestMutualTLSAuthentication
func TestMutualTLSAuthentication(t *testing.T) {
	// define certificate and private key filenames
	serverCert := CertFile(filepath.Join("cert", "serverCert.pem"))
	serverKey := KeyFile(filepath.Join("cert", "serverKey.pem"))
	clientCert := CertFile(filepath.Join("cert", "clientCert.pem"))
	clientKey := KeyFile(filepath.Join("cert", "clientKey.pem"))

	// init server-side certificate filenames
	serverCerts := CertsEx{
		Certs{
			serverCert,
			serverKey,
		},
		clientCert,
	}

	// init client-side certificate filenames
	clientCerts := CertsEx{
		Certs{
			clientCert,
			clientKey,
		},
		serverCert,
	}

	// create a context to stop the server
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	// server address
	addr := "localhost:44443"

	// create the server
	server, err := createTLSServer(ctx, addr, serverCerts)
	if err != nil {
		t.Fatal(err)
	}

	// create a channel to wait for the server to finish
	done := make(chan struct{})
	// run server asynchronously
	go runTLSServer(server, serverCerts.Certs, done, t)
	// wait until the server is ready
	server.Ready()

	// run the client
	err = runTLSClient(addr, clientCerts)
	if err != nil {
		t.Fatal(err)
	}
}

// helper types for destinguishing between certificate and key filenames
type (
	CertFile string
	KeyFile  string
)

// helper struct for storing certificate and private key
type Certs struct {
	cert CertFile
	key  KeyFile
}

// helper struct for storing certificate and private key +
// other side's certificate
type CertsEx struct {
	Certs
	peerCert CertFile
}

// func createTLSServer
func createTLSServer(
	ctx context.Context,
	addr string,
	certs CertsEx,
) (*Server, error) {
	// create TLS config
	config, err := createServerTLSConfig(certs)
	if err != nil {
		return nil, fmt.Errorf("create server TLS config: %v", err)
	}

	// create server
	server := NewTLSServer(ctx, addr, 0, config)
	return server, nil
}

// func createServerTLSConfig
func createServerTLSConfig(certs CertsEx) (*tls.Config, error) {
	// create a pool of client certificates the server trusts
	pool, err := caCertPool(certs.peerCert)
	if err != nil {
		return nil, fmt.Errorf("create certificate pool: %v", err)
	}

	// load server certificate
	cert, err := tls.LoadX509KeyPair(
		string(certs.cert),
		string(certs.key),
	)
	if err != nil {
		return nil, fmt.Errorf("load certificate: %v", err)
	}

	// create the config
	config := &tls.Config{
		// specify server's certificates
		Certificates: []tls.Certificate{cert},
		// add custom verification function to check that the client's IP
		// matches those specified in the pool's certificates
		GetConfigForClient: getConfigForClientFunc(cert, pool),
	}

	return config, nil
}

// func createClientTLSConfig
func createClientTLSConfig(certs CertsEx) (*tls.Config, error) {
	// create a pool of server certificates the client trusts
	pool, err := caCertPool(certs.peerCert)
	if err != nil {
		return nil, fmt.Errorf("create certificate pool: %v", err)
	}

	// load client certificate
	cert, err := tls.LoadX509KeyPair(
		string(certs.cert),
		string(certs.key),
	)
	if err != nil {
		return nil, fmt.Errorf("load certificate: %v", err)
	}

	// create the config
	config := &tls.Config{
		// specify client's certificates to present to the server upon request
		Certificates: []tls.Certificate{cert},
		// specify key-exchange algorithm
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
		},
		// TLS 1.3 minimum
		MinVersion: tls.VersionTLS13,
		// restrict trusted CAs to those from the create pool
		RootCAs: pool,
	}

	return config, nil
}

// func caCertPool
func caCertPool(certFile CertFile) (*x509.CertPool, error) {
	// read the certificate from the file
	cert, err := os.ReadFile(string(certFile))
	if err != nil {
		return nil, fmt.Errorf(
			"read file %q: %v",
			string(certFile),
			err,
		)
	}

	// create a certificate pool
	pool := x509.NewCertPool()

	// add the certificate to the pool
	if !pool.AppendCertsFromPEM(cert) {
		return nil, fmt.Errorf("failed to add certificate to pool")
	}

	return pool, nil
}

// func runTLSServer
func runTLSServer(
	server *Server,
	certs Certs,
	done chan struct{},
	t *testing.T,
) {
	// helper function to skip shutdown-related errors
	filterErrors := func(err error) error {
		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "use of closed network connection") {
			return nil
		}

		return err
	}

	// listen and serve, using provided certificate and private key
	err := server.ListenAndServeTLS(
		string(certs.cert),
		string(certs.key),
	)
	if filterErrors(err) != nil {
		t.Error(err)
	}

	// notify that the server finished
	done <- struct{}{}
}

// func runTLSClient
func runTLSClient(addr string, certs CertsEx) error {
	// create TLS config
	config, err := createClientTLSConfig(certs)
	if err != nil {
		return fmt.Errorf("create client TLS config: %v", err)
	}

	// established secure connection
	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("dial %s: %v", addr, err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// send hello
	hello := []byte("hello")
	_, err = conn.Write(hello)
	if err != nil {
		return fmt.Errorf("write hello: %v", err)
	}

	// read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("read hello: %v", err)
	}

	// check that the response is hello
	if actual := buf[:n]; !bytes.Equal(actual, hello) {
		return fmt.Errorf("expected %q; actual %q", hello, actual)
	}

	return nil
}

// func getConfigForClientFunc
func getConfigForClientFunc(
	cert tls.Certificate,
	pool *x509.CertPool,
) func(*tls.ClientHelloInfo) (*tls.Config, error) {
	// wrap config getter, capturing the server certificate and pool
	return func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
		return getConfigForClient(cert, pool, hello)
	}
}

// func getConfigForClient
//
// the function returns a TLS configuration specific to a client:
// although, we always return the same config,
// it could be different from client to client in other cases
func getConfigForClient(
	cert tls.Certificate,
	pool *x509.CertPool,
	hello *tls.ClientHelloInfo,
) (*tls.Config, error) {
	// the config:
	config := &tls.Config{
		// explicitly specify server certificates
		// used during the handshake with the current client
		Certificates: []tls.Certificate{cert},
		// require client certificate and its verification
		ClientAuth: tls.RequireAndVerifyClientCert,
		// specify client cerfificate verification procedure
		VerifyPeerCertificate: verifyClientCertificateFunc(pool, hello),
		// specify allowed certificates to verify the client against
		ClientCAs: pool,
		// prefer P-256 elliptic curve for key-exchange
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		// TLS version 1.3 minimum
		MinVersion: tls.VersionTLS13,
		// prefer server-side ciphers
		PreferServerCipherSuites: true,
	}

	return config, nil
}

// func verifyClientCertificateFunc
func verifyClientCertificateFunc(
	pool *x509.CertPool,
	hello *tls.ClientHelloInfo,
) func([][]byte, [][]*x509.Certificate) error {
	// wrap verification function, capturing the server's pool,
	// to check client certificates against it
	return func(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
		return verifyClientCertificate(verifiedChains, pool, hello)
	}
}

// func verifyClientCertificate
func verifyClientCertificate(
	verifiedChains [][]*x509.Certificate,
	pool *x509.CertPool,
	hello *tls.ClientHelloInfo,
) error {
	// specify usage: client authentication
	keyUsages := []x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth,
	}

	// create verification options: server pool + usage
	opts := x509.VerifyOptions{
		KeyUsages: keyUsages,
		Roots:     pool,
	}

	// get client hostnames
	hostnames, err := getHostnames(hello.Conn)
	if err != nil {
		return fmt.Errorf(
			"get hostnames for client %s: %v",
			hello.Conn.RemoteAddr().String(),
			err,
		)
	}

	// verify client certificates against the verification options
	return verifyCertificate(verifiedChains, hostnames, opts)
}

// func getHostnames
func getHostnames(conn net.Conn) ([]string, error) {
	// retrieve the client's IP address
	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
	// retrieve the client's hostnames using reverse DNS lookup
	hostnames, err := net.LookupAddr(ip)
	if err != nil {
		return nil, fmt.Errorf("lookup IP %s: %v", ip, err)
	}

	// append the IP address to the hostname list
	hostnames = append(hostnames, ip)
	return hostnames, nil
}

// func verifyCertificate
func verifyCertificate(
	verifiedChains [][]*x509.Certificate,
	hostnames []string,
	opts x509.VerifyOptions,
) error {
	// iterate through certificate chains
	// when at least one chain passes verification, we're done
	for _, chain := range verifiedChains {
		// intermediate certificates are "parents"
		// used to sign the following certificate, the root is at the end
		intermediateCerts := chain[1:]
		// build a pool of intermediate certificates (might be empty)
		opts.Intermediates = x509.NewCertPool()
		for _, cert := range intermediateCerts {
			opts.Intermediates.AddCert(cert)
		}

		// iterate through the hostname list
		for _, hostname := range hostnames {
			// set the DNS name for the verification options to verify
			// that the leaf certificate matches at least one of those
			opts.DNSName = hostname
			// leaf certificate
			leafCert := chain[0]
			// verify the certificate against the options
			_, err := leafCert.Verify(opts)
			if err == nil {
				return nil
			}
		}
	}

	// if none of the certificates succeeded verification failed
	return fmt.Errorf("client authentication failed")
}
