.PHONY: build test test-v lint clean

BIN := bin/slurm-linter

build:
	go build -o $(BIN) ./cmd/slurm-linter

test:
	go test ./...

test-v:
	go test -v ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
