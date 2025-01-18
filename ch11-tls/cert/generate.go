package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

// CLI options
var (
	// host names (IPs and DNS names)
	hosts = flag.String(
		"hosts",
		"localhost",
		"certificate's comma-separated host names and IPs",
	)

	// certificate filename
	certFilename = flag.String(
		"cert",
		"cert.pem",
		"certificate file name",
	)

	// private key filename
	privKeyFilename = flag.String(
		"key",
		"key.pem",
		"private key file name",
	)
)

func main() {
	// parse command line arguments
	flag.Parse()

	// split host string
	hostsSplitted := strings.Split(*hosts, ",")
	// create certificate template
	template, err := createTemplate(hostsSplitted)
	if err != nil {
		log.Fatalf("create template: %v", err)
	}

	// generate certificate + private key
	cert, privKey, err := generateCertAndPrivKey(template)
	if err != nil {
		log.Fatalf("generate certificate: %v", err)
	}

	// save the certificate to the specified file
	err = saveCertificate(cert, *certFilename)
	if err != nil {
		log.Fatalf("save certificate: %v", err)
	} else {
		log.Printf("wrote %s\n", *certFilename)
	}

	// save the private key to the specified file
	err = savePrivKey(privKey, *privKeyFilename)
	if err != nil {
		log.Fatalf("save private key: %v", err)
	} else {
		log.Printf("wrote %s\n", *privKeyFilename)
	}
}

// func createTemplate
func createTemplate(hosts []string) (*x509.Certificate, error) {
	// upper limit for generating a random serial number
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	// generate a random serial number
	serial, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("generate serial number: %v", err)
	}

	// validity period bounds
	notBefore := time.Now()
	validityPeriod := 10 * 365 * 24 * time.Hour
	notAfter := notBefore.Add(validityPeriod)

	// usage: key exchange and encryption, signing data, signing certificates
	usage := x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDigitalSignature |
		x509.KeyUsageCertSign

	// addtional usage scenarios: server and client authentication
	extUsage := []x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth,
		x509.ExtKeyUsageServerAuth,
	}

	// specify the issuer
	subject := pkix.Name{
		Organization: []string{"Dmitrii Kochetov"},
	}

	// create a new template
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      subject,
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     usage,
		ExtKeyUsage:  extUsage,
		// basic consttraints are valid
		BasicConstraintsValid: true,
		// this certificate is CA (can be used to sign certificates)
		IsCA: true,
	}

	// add IPs and host names to the template
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}

	return &template, nil
}

// func generateCertAndPrivKey
func generateCertAndPrivKey(template *x509.Certificate) (
	[]byte, // certificate
	[]byte, // private key
	error,
) {
	// generate a private key using P-256 elliptic curve
	privKey, err := ecdsa.GenerateKey(
		elliptic.P256(),
		rand.Reader,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("generate private key: %v", err)
	}

	// create an x509 certificate using the provided template
	// template is used for both subject and issuer
	// (self-signed certificate)
	//
	// here we generate the certificate based on the generated private key
	// and embed the corresponding public key into it
	//
	// the certificate is in DER format
	// (binary encoding of X509 certificates)
	derCert, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		&privKey.PublicKey,
		privKey,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create certificate: %v", err)
	}

	// convert the private key into the PKCS#8 format,
	// a standardized way of storing private keys
	pkcs8privKey, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("convert private key to PKCS#8: %v", err)
	}

	return derCert, pkcs8privKey, nil
}

// func saveCertificate
func saveCertificate(cert []byte, filename string) error {
	// create/open the file for writing (truncate if exists)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file %q: %v", filename, err)
	}
	// close the file at scope exit
	defer func() {
		_ = file.Close()
	}()

	// write the DER encoded certificate in PEM-format,
	// human-readable text encoding format with a header and a footer
	err = pem.Encode(
		file,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		},
	)
	if err != nil {
		return fmt.Errorf("encode certificate: %v", err)
	}

	return nil
}

// func savePrivKey
func savePrivKey(privKey []byte, filename string) error {
	// open/create a file to write the private key to
	// (same as in saveCertificate, but with access restrictions)
	file, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return fmt.Errorf("create file %q: %v", filename, err)
	}
	defer func() {
		_ = file.Close()
	}()

	// write the private key encoded in PKCS#8 format in PEM-format
	err = pem.Encode(
		file,
		&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: privKey,
		},
	)
	if err != nil {
		return fmt.Errorf("encode private key: %v", err)
	}

	return nil
}
