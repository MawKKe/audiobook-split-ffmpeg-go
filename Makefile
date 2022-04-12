.DEFAULT_GOAL := exe

build:
	go build ./...

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

# the fact that this is even needed is idiotic but whatchugonnado #golang #fuckyouweknowbetter
fix:
	find . -type f -iname "*.go" -exec goimports -w {} +

exe:
	go build ./cmd/audiobook-split-ffmpeg-go

clean:
	go clean -x ./...

.PHONY: build test fmt vet clean
