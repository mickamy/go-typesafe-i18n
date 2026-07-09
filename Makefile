.PHONY: all build clean test lint generate check-generate

all: build

build:
	go build ./...
	cd examples/basic && go build ./...

clean:
	rm -rf bin

test:
	go test ./... -race
	cd examples/basic && go test ./... && go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is not installed"; \
		exit 1; \
	}
	golangci-lint run
	cd examples/basic && golangci-lint run

generate:
	cd examples/basic && go generate ./...

check-generate: generate
	git diff --exit-code examples/basic/messages
