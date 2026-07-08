.PHONY: all build clean test lint

all: build

build:
	go build ./...

clean:
	rm -rf bin

test:
	go test ./... -race

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is not installed"; \
		exit 1; \
	}
	golangci-lint run
