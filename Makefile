BINARY  := fogelbot
GOFLAGS := -ldflags="-s -w"

.PHONY: build build-alpine test test-race clean

build:
	go build $(GOFLAGS) -o $(BINARY) .

build-alpine:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o $(BINARY) .

test:
	go test ./...

test-race:
	go test -race ./...

clean:
	rm -f $(BINARY)
