package main

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/net/context/ctxhttp"
)

var (
	// Version is the version of the application
	Version = "DEV-SNAPSHOT"
)

func main() {
	// Check if we were asked the version information.
	for _, arg := range os.Args[1:] {
		if arg == "version" || arg == "-v" || arg == "-version" {
			fmt.Printf("%v %v\n", filepath.Base(os.Args[0]), Version)
			return
		}
	}

	parseArgs()

	getCertificate()
}

// getCertificate handles registering the certificates.
//
// It does the registration with the ACME directory, extracts the domains from the CSR,
// asks for the challenges, authorizes the domains, and finally gets the signed certificate.
func getCertificate() {
	account := &acme.Account{AgreedTerms: Config.AccountTerms}
	if Config.AccountEmail != "" {
		account.Contact = []string{"mailto:" + Config.AccountEmail}
	}
	client := &acme.Client{
		Key:          Config.AccountKey,
		DirectoryURL: Config.Server.String(),
	}

	ctx := context.Background()

	// We default to false here because we want people to update the command or
	// the binary when the terms change (function will be called if the terms
	// change and we don't update the terms we accept). If we auto accept, they
	// can end up agreeing to terms they didn't want.
	if a, err := client.Register(ctx, account, func(tosURL string) bool { return false }); err != nil {
		if rerr, ok := err.(*acme.Error); ok && rerr.StatusCode == 409 {
			// An account with this key exists.
			location := rerr.Header.Get("Location")
			if location == "" {
				// We have a non-compliant server
				log.Fatalf("reg: server returned 409 (%v) but no location prodived", err)
			}
			// We get the absolute URL based on the existing server URL
			accountURL, uerr := Config.Server.Parse(location)
			if uerr != nil {
				log.Fatalf("reg: could not parse acme account URL %v", err)
			}
			account.URI = accountURL.String()

			// If an account exists, we update the terms and the contacts
			//if ac, uerr := client.UpdateReg(ctx, account); uerr != nil {
			//	log.Fatalf("reg: %v", uerr)
			//} else {
			//	account = ac
			//}
		} else {
			log.Fatalf("reg: %v", err)
		}
	} else {
		account = a
	}

	var domains []string

	if Config.CSR.Subject.CommonName != "" {
		domains = append(domains, Config.CSR.Subject.CommonName)
	}
	domains = append(domains, Config.CSR.DNSNames...)
	var wg sync.WaitGroup
	authCtx, authCancel := context.WithTimeout(ctx, 10*time.Minute)
	defer authCancel()

	for _, d := range domains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			err := authorize(authCtx, client, domain)
			if err != nil {
				authCancel()
				log.Fatalf("challenge (%v): %v", domain, err)
			}
		}(d)
	}
	wg.Wait()

	certCtx, certCancel := context.WithTimeout(ctx, 30*time.Minute)
	defer certCancel()

	// At this point, all of the challenges have succeeded. Yay!
	// We send 0 as an expiration because we don't want to send "not-after"
	// attribute.
	cert, _, err := client.CreateCert(certCtx, Config.CSR.Raw, 0, true)
	if err != nil {
		log.Fatalf("cert: %v", err)
	}
	var pemcert []byte
	for _, b := range cert {
		b = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b})
		pemcert = append(pemcert, b...)
	}
	if err := ioutil.WriteFile(Config.CertificatePath, pemcert, 0644); err != nil {
		log.Fatalf("cert: %v", err)
	}
}

// authorize accepts the http-01 challenge, generates the corresponding response, and retrieves the authorization.
func authorize(ctx context.Context, client *acme.Client, domain string) error {
	// Get the challenges for the domain.
	authz, err := client.Authorize(ctx, domain)
	if err != nil {
		return err
	}

	// If we are valid, we can return early. Yay!
	if authz.Status == acme.StatusValid {
		return nil
	}

	// We only support http-01, so we find the first of that type.
	var chal *acme.Challenge
	for _, c := range authz.Challenges {
		if c.Type == "http-01" {
			chal = c
			break
		}
	}
	if chal == nil {
		return errors.New("no supported challenge found")
	}

	// We write the challenge to the file.
	response, err := client.HTTP01ChallengeResponse(chal.Token)
	if err != nil {
		return fmt.Errorf("could not generate the challenge response: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(Config.ChallengeDir, chal.Token), []byte(response), 0644)
	if err != nil {
		return fmt.Errorf("could not output challenge response: %v", err)
	}
	//noinspection GoUnhandledErrorResult
	defer os.Remove(filepath.Join(Config.ChallengeDir, chal.Token))

	if !Config.SkipSelfCheck {
		// We check that we can access it before telling ACME that it's all good.
		// HTTP01ChallengePath prefixes with /, so we don't add one.
		url := "http://" + domain + client.HTTP01ChallengePath(chal.Token)
		res, err := ctxhttp.Get(ctx, http.DefaultClient, url)
		if err != nil {
			return err
		}
		if res.StatusCode < 200 || res.StatusCode > 299 {
			return fmt.Errorf("StatusCode %d: could not read authentication at %v", res.StatusCode, url)
		}
	}

	// We tell ACME that we accept the challenge and are ready for verification.
	if _, err = client.Accept(ctx, chal); err != nil {
		return fmt.Errorf("could not accept challenge: %v", err)
	}

	_, err = client.WaitAuthorization(ctx, authz.URI)
	return err
}
