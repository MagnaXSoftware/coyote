coyote
======

[![build status](https://git.magnax.ca/magnax/coyote/badges/master/build.svg)](https://git.magnax.ca/magnax/coyote/builds)

Canonical Source
----------------

The canonical source for this project is on the MagnaX Software git server. It
may be duplicated elsewhere, but in case of doubt, go to
[MagnaX GitLab](https://git.magnax.ca/magnax/coyote).

About
-----

This tool is a restricted (reduced functionality) ACME client. It implements
portion of the ACME client spec to allow for account registration and
certificate issuance. Other aspects of the spec, such as certificate revocation,
have not been implemented to simplify the application.

The tool is inspired by the [acme-tiny](https://github.com/diafygi/acme-tiny)
script, which aims to be a tiny, auditable program to issue certificates. Like
acme-tiny, this tool does not need to bind to port 80, restart your webserver,
or change your webserver config. It simply needs to be able to read the
certificate request (CSR) and write to a specific folder in the webroot and to
the file that will hold the generated certificate.

Once compiled, this program has **no external dependency**. This makes it easy
to install on a variety of platforms. It is also very small, which makes it easy
to read and audit prior to usage (recommended).

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
| cert          | **required** | The path to the certificate file. This file will be overridden if the certificate is successfully issued. The chain will be appended after the certificate.                                        |

How To
------

Steps 0-3 are setup actions that only need to be completed once.

The following steps use openssl to generate the various components, but any tool
can be used, as long as the resulting files are in the correct format (PEM).

### 0: Setup your webserver

The challenge we are responding to (`http-01`) needs to load a specific file
over http. Given that we don't want to reconfigure or stop the webserver every
time, we'll configure a permanent challenge directory.

The files must be served from the `.well-known/acme-challenge/` directory, and
the webserver must respond over http on port 80 (because the ACME spec requires
it).

#### NGINX

Create a directory to hold the challenge files:

    mkdir -p /var/www/challenges/

Configure nginx to serve the well-known folder from that directory.

    server {
        listen 80;
        server_name yoursite.com www.yoursite.com;

        location /.well-known/acme-challenge/ {
            alias /var/www/challenges/;
            try_files $uri =404;
        }

        ...the rest of your config
    }

### 1: Create an account key

Skip if you already have an account key in PEM format. If you have a key in JWK
format (like the official client generates), you'll have to convert it first.

You must have a public key registered with the ACME server. This keypair is used
to sign communications and requests between server/clients. You can use any tool
that can generate an RSA key to create your account key.

Ensure that this file is well protected. Only coyote needs to be able to read
it.

    openssl genrsa 4096 > account.key

### 2: Create a certificate private key

This keypair will be the one used in the certificate. Keep it secure.

    openssl genrsa 4096 > domain.key

### 3: Generate a certificate request (CSR)

Coyote will read the CSR to determine which domains to authorize. Make sure that
coyote will be able to read the file. The CSR needs to not be protected by a
password, as coyote has no support for that at the moment.

#### Single domain

    openssl req -new -sha256 -key domain.key -subj "/CN=example.com" > domain.csr

#### Multi domain

    openssl req -new -sha256 -key domain.key -subj "/" -reqexts SAN -config <(cat /etc/ssl/openssl.cnf <(printf "[SAN]\nsubjectAltName=DNS:example.com,DNS:www.example.com")) > domain.csr

You can change the path to the openssl.cnf file depending on your environment.

### 4: Generate the certificate

Everything is now setup for you to be able to generate the certificate. Coyote
will run, generate the certificate, and exit. If an error occurs, it will
provide output (silent means that all went well, so great for cron!) and return
a non-zero number. The `domain.crt` file will contain the full certificate
chain, starting with the requested certificate (this means that the chain will
always be correct for the generated certificate).

The following command can be re-run as needed (within the limits of the ACME
server) to re-generate/renew the certificates.

    coyote -account-key account.key -challenge-dir /var/www/challenges -csr domain.csr -cert domain.crt

### 5: Install the certificate

This part depends on which server is using certificate, but typically you'll
give it the path to the `domain.key` and `domain.crt` files.

A lot of server software is capable of reading both the certificate and the
chain from the same file, so that simplifies configuration substantially.

### 6: Setup a cronjob for autorenewal

One of the advantages of ACME in combination with clients such as coyote is that
the renewal process can be entirely automated. Certificates will be re-generated
on a schedule and the server will be restarted to pick up the new certificate
automatically.

For instance, you can setup a renewal for nginx as follow:

    0 0 2 */2 * coyote -account-key account.key -challenge-dir /var/www/challenges -csr domain.csr -cert domain.crt && service nginx reload
