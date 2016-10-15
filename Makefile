ARCH ?= amd64 386 arm
OS ?= !openbsd !netbsd !plan9

all: tools fmt build

tag:
	git tag -a -s -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}


clean:
	@rm -rf ./build
	@rm -f ./coyote

tools:
	go get -u github.com/kardianos/govendor
	go get -u github.com/mitchellh/gox

deps: tools
	govendor sync

fmt: deps
	go fmt .

vet: deps
	go vet -v .

build: deps
	go build .

release: tools deps clean
	gox -os="${OS}" -arch="${ARCH}" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" .

.PNONY: all tag clean tools deps fmt vet build release
