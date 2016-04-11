package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	acme "github.com/google/goacme"
)

// Config represents the configuration options for coyote.
var Config struct {
	Server          acme.Endpoint
	AccountKey      *rsa.PrivateKey
	AccountEmail    string
	AccountTerms    string
	ChallengeDir    string
	CSR             *x509.CertificateRequest
	CertificatePath string
}

const (
	rsaPrivateKey = "RSA PRIVATE KEY"
	x509CSR       = "CERTIFICATE REQUEST"
)

func init() {
	acmeServerURL := flag.String("acme-server", "https://acme-v01.api.letsencrypt.org/directory", "URL of the ACME server directory. defaults to the Let's Encrypt live server")
	accountKeyPath := flag.String("account-key", "", "Path to your Let's Encrypt account private key.")
	flag.StringVar(&Config.AccountEmail, "account-email", "", "The email to associate with the registration.")
	flag.StringVar(&Config.AccountTerms, "account-terms", "https://letsencrypt.org/documents/LE-SA-v1.0.1-July-27-2015.pdf", "The terms that need to be accepted before certificate issuance.")
	flag.StringVar(&Config.ChallengeDir, "challenge-dir", ".well-known/acme-challenge/", "Path to the challenge directory.")
	csrPath := flag.String("csr", "", "Path to your CSR file.")
	flag.StringVar(&Config.CertificatePath, "cert", "", "Path to the certificate file (with chain)")

	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			flag.Usage()
			fmt.Fprintln(os.Stderr, r)
			os.Exit(1)
		}
	}()

	var err error

	Config.Server, err = acme.Discover(nil, *acmeServerURL)
	if err != nil {
		panic(err)
	}

	if *accountKeyPath == "" {
		panic("no account key supplied")
	}
	privKey, err := readKey(*accountKeyPath)
	if err != nil {
		panic(err)
	}
	Config.AccountKey = privKey

	if *csrPath == "" {
		panic("no CSR supplied")
	}
	csr, err := readCSR(*csrPath)
	if err != nil {
		panic(err)
	}
	Config.CSR = csr

	if _, err := os.Stat(Config.ChallengeDir); err != nil {
		panic(fmt.Sprintf("can't access %q", Config.ChallengeDir))
	}

	if Config.CertificatePath == "" {
		panic("no Certificate Path supplied")
	}
	Config.CertificatePath = filepath.Clean(Config.CertificatePath)
}

func readKey(path string) (*rsa.PrivateKey, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode(raw)
	if data == nil {
		return nil, fmt.Errorf("no block found in %q", path)
	}
	if data.Type != rsaPrivateKey {
		return nil, fmt.Errorf("%q is unsupported", data.Type)
	}
	return x509.ParsePKCS1PrivateKey(data.Bytes)
}

func readCSR(path string) (*x509.CertificateRequest, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode(raw)
	if data == nil {
		return nil, fmt.Errorf("no block found in %q", path)
	}
	if data.Type != x509CSR {
		return nil, fmt.Errorf("%q is unsupported", data.Type)
	}

	return x509.ParseCertificateRequest(data.Bytes)
}