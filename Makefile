VERSION=0.1.1

all: deps fmt combined

combined:
		go install .

release: release-deps
		gox -os="!freebsd !openbsd !netbsd !plan9" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" .

fmt:
		go fmt ./...

deps:
		go get github.com/aws/aws-sdk-go

release-deps:
		go get github.com/mitchellh/gox

pull:
		git pull

tag:
		git tag -a -s -m 'v${VERSION}' ${VERSION} && git push origin ${VERSION}

.PNONY: all combined release fmt deps release-deps

