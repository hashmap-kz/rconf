.PHONY: lint run build test

lint:
	golangci-lint run ./...

run:
	go run ./main.go

build:
	go build ./main.go

test:
	go test -cover ./...

