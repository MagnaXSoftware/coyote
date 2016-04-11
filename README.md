coyote
======

[![build status](https://git.magnax.ca/magnax/coyote/badges/master/build.svg)](https://git.magnax.ca/magnax/coyote/builds)

Canonical Source
----------------

The canonical source for this project is on the MagnaX Software git server. It
may be duplicated elsewhere, but in case of doubt, go to
git.magnax.ca/magnax/coyote.

About
-----

This tool is a restricted ACME client. It implements portion of the ACME client
spec to allow for account registration and certificate issuance. Other aspects
of the spec, such as certificate revokation, have not been implemented to
simplify the implementation.

The tool is inspired by the [acme-tiny](https://github.com/diafygi/acme-tiny)
script, which aims to be a tiny, auditable script to issue certificate. It
doesn't need to bind to port 80, restart your webserver, or change your
webserver config. This tool doesn't either. However it needs to be able to write
files that the webserver needs to be configured to serve.

Usage
-----

The tool has a few command line arguments:

| name          | status       | usage                                                                                                                                                                                               |
|---------------|--------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| account-email | optional     | The email to associate with the account for things like recovery purposes.                                                                                                                          |
| account-key   | **required** | The path to the private key used for the account. Try not to change it.                                                                                                                             |
| account-terms | optional     | The terms that should be auto-accepted before certificate issuance. Make sure that you read them, because coyote auto-accepts them. They default to the let's encrypt terms from July 27th 2015.    |
| acme-server   | optional     | The URL of the ACME server directory. The directory returns a json object listing the url of the ACME endpoints. The default is the let's encrypt live v01 server.                                  |
| challenge-dir | **required** | The directory that will contain the http challenges that will be served by the webserver. Make sure that coyote can write to it, and that the browser serves it from _.well-known/acme-challenge/_. |
| csr           | **required** | The path to the CSR with the domains to issue a certificate for. Subject Alternative Names are accepted.                                                                                            |
| cert          | **required** | The path to the certificate file. This file will be overridden if the certificate is succcessfully issued. The chain will be appended after the certificate.                                        |
