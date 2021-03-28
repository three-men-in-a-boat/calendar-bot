PROJECT=calendar-bot
ORGANISATION=three-men-in-a-boat
SOURCE=$(shell find . -name '*.go' | grep -v vendor/)
SOURCE_DIRS = cmd pkg

export GO111MODULE=on

.PHONY: vendor vetcheck fmtcheck clean build build-debug gotest

all: vendor vetcheck fmtcheck build gotest mod-clean

debug: vendor vetcheck fmtcheck build-debug gotest mod-clean

ver:
	@echo Building version: $(VERSION)

build: $(SOURCE)
	@mkdir -p build/bin
	go build -o build/bin/botbackend ./cmd/main.go

build-linux-amd64:
	@CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o build/bin/linux-amd64/botbackend ./cmd/main.go

gotest:
	go test -cover ./...

fmtcheck:
	@gofmt -l -s $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

mod-clean:
	go mod tidy

clean:
	@rm -rf build
	go mod tidy

vendor:
	go mod vendor

vetcheck:
	go vet ./...
	golangci-lint run

build-debug:
	@mkdir -p build/debug
	go build -o build/debug/botbackend -gcflags="all=-N -l" ./cmd/main.go

build-debug-linux-amd64:
	@mkdir -p build/debug/linux-amd64
	go build -o build/debug/linux-amd64/botbackend -gcflags="all=-N -l" ./cmd/main.go
