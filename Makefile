GitCommit := $(shell git rev-parse --short=8 HEAD)
LDFLAGS := "-s -w -X main.GitCommit=$(GitCommit)" -trimpath

build:
	@go build -ldflags $(LDFLAGS) -o ./bin/gmdump ./cmd/gmdump

clean:
	@rm -rf ./bin ./release

release:
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o ./release/gmdump-darwin-amd64 ./cmd/gmdump
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o ./release/gmdump-linux-amd64 ./cmd/gmdump
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS) -o ./release/gmdump-windows-amd64.exe ./cmd/gmdump

.PHONY: build clean release
