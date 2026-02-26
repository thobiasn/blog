.PHONY: dev build test clean

dev:
	go run ./cmd/blog serve

build:
	go build -o blog ./cmd/blog

test:
	go test ./...

clean:
	rm -f blog
