package main

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

// Config represents the configuration options for coyote.
var Config struct {
	Server          *url.URL
	AccountKey      crypto.Signer
	AccountEmail    string
	AccountTerms    string
	ChallengeDir    string
	CSR             *x509.CertificateRequest
	CertificatePath string
}

const (
	rsaPrivateKey = "RSA PRIVATE KEY"
	ecPrivateKey  = "EC PRIVATE KEY"
	x509CSR       = "CERTIFICATE REQUEST"
)

var (
	acmeServerURL  string
	accountKeyPath string
	csrPath        string
)

// init sets up the flags for the configuration.
func init() {
	flag.StringVar(&acmeServerURL, "acme-server", "https://acme-v01.api.letsencrypt.org/directory", "URL of the ACME server directory. defaults to the Let's Encrypt live server")
	flag.StringVar(&accountKeyPath, "account-key", "", "Path to your Let's Encrypt account private key.")
	flag.StringVar(&Config.AccountEmail, "account-email", "", "The email to associate with the registration.")
	flag.StringVar(&Config.AccountTerms, "account-terms", "https://letsencrypt.org/documents/LE-SA-v1.1.1-August-1-2016.pdf", "The terms that need to be accepted before certificate issuance.")
	flag.StringVar(&Config.ChallengeDir, "challenge-dir", ".well-known/acme-challenge/", "Path to the challenge directory.")
	flag.StringVar(&csrPath, "csr", "", "Path to your CSR file.")
	flag.StringVar(&Config.CertificatePath, "cert", "", "Path to the certificate file (with chain)")
}

// parseArgs is called to actually populate the Config structure with the args.
func parseArgs() {
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			flag.Usage()
			fmt.Fprintln(os.Stderr, r)
			os.Exit(1)
		}
	}()

	var err error

	if acmeServerURL == "" {
		panic("no acme discovery URL provided")
	}
	Config.Server, err = url.Parse(acmeServerURL)
	if err != nil {
		panic(err)
	}

	if accountKeyPath == "" {
		panic("no account key supplied")
	}
	privKey, err := readKey(accountKeyPath)
	if err != nil {
		panic(err)
	}
	Config.AccountKey = privKey

	if csrPath == "" {
		panic("no CSR supplied")
	}
	csr, err := readCSR(csrPath)
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

// readKey reads a private key from a given path.
// The key is expected to be PEM encoded.
func readKey(path string) (crypto.Signer, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode(raw)
	if data == nil {
		return nil, fmt.Errorf("no block found in %q", path)
	}
	switch data.Type {
	case rsaPrivateKey:
		return x509.ParsePKCS1PrivateKey(data.Bytes)
	case ecPrivateKey:
		return x509.ParseECPrivateKey(data.Bytes)
	default:
		return nil, fmt.Errorf("%q is unsupported", data.Type)
	}
}

// readCSR reads a certificate request from a given path.
// The request is expected to be PEM encoded.
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
