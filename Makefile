.DEFAULT_GOAL := audiobook-split-ffmpeg

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

audiobook-split-ffmpeg:
	go build -o $@ ./cmd/audiobook-split-ffmpeg

.PHONY: build test fmt vet audiobook-split-ffmpeg
