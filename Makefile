BUILD_ARCH ?= amd64 386 arm
BUILD_OS ?= !netbsd !plan9
BUILD_VERSION ?= $(shell git describe --long --dirty)

all: tools clean fmt build

tag:
	git tag -a -s -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}


clean:
	@rm -rf ./build
	@rm -f ./coyote
	@rm -f ./coyote.exe

fmt:
	go fmt .

vet:
	go vet -v .

build:
	go build -ldflags "-X main.Version=${BUILD_VERSION}" .

tools:
	go get -u github.com/mitchellh/gox

release: tools clean
	gox -os="${BUILD_OS}" -arch="${BUILD_ARCH}" -ldflags "-X main.Version=${BUILD_VERSION}" -output="build/{{.Dir}}-{{.OS}}-{{.Arch}}" .

.PNONY: all tag clean tools fmt vet build release
