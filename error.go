package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Predefined Error.Type values by the ACME spec.
const (
	ErrBadCSR       = "urn:acme:error:badCSR"
	ErrBadNonce     = "urn:acme:error:badNonce"
	ErrConnection   = "urn:acme:error:connection"
	ErrDNSSec       = "urn:acme:error:dnssec"
	ErrMalformed    = "urn:acme:error:malformed"
	ErrInternal     = "urn:acme:error:serverInternal"
	ErrTLS          = "urn:acme:error:tls"
	ErrUnauthorized = "urn:acme:error:unauthorized"
	ErrUnknownHost  = "urn:acme:error:unknownHost"
	ErrRateLimited  = "urn:acme:error:rateLimited"
)

// ACMEError is an ACME error.
type ACMEError struct {
	Status int
	Type   string
	Detail string
}

func (e *ACMEError) Error() string {
	return fmt.Sprintf("%d %s: %s", e.Status, e.Type, e.Detail)
}

// responseError creates an error of Error type from resp.
func responseError(resp *http.Response) error {
	// don't care if ReadAll returns an error:
	// json.Unmarshal will fail in that case anyway
	b, _ := ioutil.ReadAll(resp.Body)
	e := &ACMEError{Status: resp.StatusCode}
	if err := json.Unmarshal(b, e); err != nil {
		// so the body does not contain a json error message. Try something else.
		e.Detail = string(b)
		if e.Detail == "" {
			e.Detail = resp.Status
		}
	}
	return e
}

// TemporaryError is a "temporary" error indicating that the request
// should be retried after the specified duration.
type TemporaryError time.Duration

func (re TemporaryError) Error() string {
	return fmt.Sprintf("retry after %s", re)
}
