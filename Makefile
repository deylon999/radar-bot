.PHONY: run build test lint

run:
	go run ./cmd/bot

build:
	go build -o bin/bot ./cmd/bot

test:
	go test ./...

lint:
	go vet ./...
	test -z "$$(gofmt -l .)"
