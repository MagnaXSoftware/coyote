all: fmt combined

combined:
	go install .

release: deps release-deps
	gox -os="!openbsd !netbsd !plan9" -arch="!arm" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" .

fmt:
	go fmt .

deps:
	go get ./...

release-deps:
	go get github.com/mitchellh/gox

pull:
	git pull

tag:
	git tag -a -s -m 'v${VERSION}' ${VERSION} && git push origin ${VERSION}

build:
	go build ./...

.PNONY: all combined release fmt deps release-deps build
