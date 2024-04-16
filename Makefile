all:
	go install github.com/mike1808/ax/cmd/ax

test:
	go test ./cmd/... ./pkg/...

deps:
	dep ensure

release:
	rm -rf dist bin
	goreleaser
