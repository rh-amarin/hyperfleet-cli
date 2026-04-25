BIN       := hf
MODULE    := github.com/rh-amarin/hyperfleet-cli
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE      := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   := -ldflags "-X $(MODULE)/internal/version.Version=$(VERSION) \
                        -X $(MODULE)/internal/version.Commit=$(COMMIT) \
                        -X $(MODULE)/internal/version.Date=$(DATE)"

.PHONY: build install lint test clean

build:
	go build $(LDFLAGS) -o bin/$(BIN) .

install:
	go install $(LDFLAGS) .

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	rm -rf bin/
