ARCH ?= "amd64 386 arm"

all: fmt combined

combined:
	go install .

release: deps release-deps
	gox -os="!openbsd !netbsd !plan9" -arch="${ARCH}" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" .

fmt:
	go fmt .

deps:
	go get -d ./...

release-deps:
	go get github.com/mitchellh/gox

pull:
	git pull

vet:
	go vet -v ./...

tag:
	git tag -a -s -m 'v${VERSION}' v${VERSION} && git push origin ${VERSION}

build: deps
	go build ./...

.PNONY: all combined release fmt deps release-deps build deps vet
