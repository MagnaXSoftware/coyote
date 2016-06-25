ARCH ?= "amd64 386 arm"

all: fmt combined

tag:
	git tag -a -s -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}

combined:
	go install .

fmt:
	go fmt .

vet:
	go vet -v ./...

deps:
	go get -d ./...

release-deps:
	go get -u github.com/mitchellh/gox

build: deps
	go build ./...

release: deps release-deps
	gox -os="!openbsd !netbsd !plan9" -arch="${ARCH}" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" .

.PNONY: all combined release fmt deps release-deps build deps vet
