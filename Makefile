PROJECT=calendar-bot
ORGANISATION=three-men-in-a-boat
SOURCE=$(shell find . -name '*.go' | grep -v vendor/)
SOURCE_DIRS = cmd pkg

export GO111MODULE=on

.PHONY: vendor vetcheck fmtcheck clean build gotest

all: vendor vetcheck fmtcheck build gotest mod-clean

ver:
	@echo Building version: $(VERSION)

build: $(SOURCE)
	@mkdir -p build/bin
	go build -o build/bin/botbackend ./cmd/main.go

gotest:
	go test -cover ./...

fmtcheck:
	@gofmt -l -s $(SOURCE_DIRS)

mod-clean:
	go mod tidy

clean:
	@rm -rf build
	go mod tidy

vendor:
	go mod vendor

vetcheck:
	go list ./... | grep -v bn254 | xargs go vet
	golangci-lint run
