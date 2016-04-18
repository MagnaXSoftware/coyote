package main

import (
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	acme "github.com/google/goacme"
)

func main() {
	account := &acme.Account{AgreedTerms: Config.AccountTerms}
	if Config.AccountEmail != "" {
		account.Contact = []string{"mailto:" + Config.AccountEmail}
	}
	client := &acme.Client{Key: Config.AccountKey}

	if err := client.Register(Config.Server.RegURL, account); err != nil {
		rerr, ok := err.(*acme.Error)
		if ok && rerr.Status == 409 {
			url, lerr := rerr.Response.Location()
			if lerr != nil {
				log.Fatalf("reg: %v", lerr)
			}

			if uerr := client.UpdateReg(url.String(), account); uerr != nil {
				log.Fatalf("reg: %v", uerr)
			}
		} else {
			log.Fatalf("reg: %v", err)
		}
	}

	var domains []string

	if Config.CSR.Subject.CommonName != "" {
		domains = append(domains, Config.CSR.Subject.CommonName)
	}
	domains = append(domains, Config.CSR.DNSNames...)
	httpClient := &http.Client{}
	var wg sync.WaitGroup

	for _, domain := range domains {
		wg.Add(1)
		go func(dom string) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					exit()
				}
			}()
			authorize(client, dom, httpClient)
		}(domain)
	}
	wg.Wait()

	// At this point, all of the challenges have succeeded. Yay!
	cert, curl, err := client.CreateCert(Config.Server.CertURL, Config.CSR.Raw, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	if cert == nil {
		cert = pollCert(curl)
	}
	var pemcert []byte
	for _, b := range cert {
		b = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b})
		pemcert = append(pemcert, b...)
	}
	if err := ioutil.WriteFile(Config.CertificatePath, pemcert, 0644); err != nil {
		log.Fatalf("write cert: %v", err)
	}
}

func authorize(client *acme.Client, domain string, httpClient *http.Client) {
	// Get the challenges for the domain.
	authz, err := client.Authorize(Config.Server.AuthzURL, domain)
	if err != nil {
		log.Panicf("challenge: %v", err)
	}
	// We only support http-01, so we find the first of that type.
	var chal *acme.Challenge
	for _, c := range authz.Challenges {
		if c.Type == "http-01" {
			chal = &c
			break
		}
	}
	if chal == nil {
		log.Panicf("challenge: no supported challenge found")
	}

	// We write the challenge to the file.
	ioutil.WriteFile(filepath.Join(Config.ChallengeDir, chal.Token), []byte(keyAuth(&Config.AccountKey.PublicKey, chal.Token)), 0644)
	defer os.Remove(filepath.Join(Config.ChallengeDir, chal.Token))

	// We check that we can access it before telling ACME that it's all good.
	url := "http://" + domain + "/.well-known/acme-challenge/" + chal.Token
	res, err := httpClient.Get(url)
	if err != nil {
		log.Panicf("check challenge: %v", err)
	}
	if res.StatusCode != 200 {
		log.Panicf("auth: Could not get authorization at %s", url)
	}

	// We tell ACME that we accept the challenge and are ready for verification.
	if _, err := client.Accept(chal); err != nil {
		log.Panicf("accept challenge: %v", err)
	}
	for {
		a, err := client.GetAuthz(authz.URI)
		if err != nil {
			log.Panicf("authz %q: %v", authz.URI, err)
		}
		if a.Status == acme.StatusInvalid {
			log.Panicf("could not get certificate for %s", domain)
		}
		if a.Status != acme.StatusValid {
			time.Sleep(time.Duration(3) * time.Second)
			continue
		}
		break
	}

}

// keyAuth generates a key authorization string for a given token.
func keyAuth(pub *rsa.PublicKey, token string) string {
	return fmt.Sprintf("%s.%s", token, acme.JWKThumbprint(pub))
}

func pollCert(url string) [][]byte {
	for {
		b, err := acme.FetchCert(nil, url, true)
		if err == nil {
			return b
		}
		d := 3 * time.Second
		if re, ok := err.(acme.RetryError); ok {
			d = time.Duration(re)
		}
		time.Sleep(d)
	}
}

func exit() {
	os.Exit(2)
}
