.DEFAULT_GOAL := exe

PROJECT_URL := github.com/MawKKe/audiobook-split-ffmpeg-go

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

git_latest_version_tag := git describe --tags --match "v[0-9]*" --abbrev=0

# Make sure the tags are published and pushed to the public remote!
sync-package-proxy:
	GOPROXY=proxy.golang.org go list -m ${PROJECT_URL}@$(shell ${git_latest_version_tag})

.PHONY: build test fmt vet clean sync-package-proxy
