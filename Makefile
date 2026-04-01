.PHONY: build test test-v lint clean coverage

BIN := bin/slurm-linter

build:
	go build -o $(BIN) ./cmd/slurm-linter

test:
	go test ./...

test-v:
	go test -v ./...

lint:
	go vet ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

clean:
	rm -rf bin/ coverage.out coverage.html
