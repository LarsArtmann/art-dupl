.PHONY: clean check test build

default: clean check test build

clean:
	rm -rf dist/ cover.out

test: clean
	GOEXPERIMENT=jsonv2 go test -v -cover ./...

check:
	GOEXPERIMENT=jsonv2 golangci-lint run

build:
	GOEXPERIMENT=jsonv2 go build -ldflags "-s -w" -trimpath
