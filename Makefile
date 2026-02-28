.PHONY: dev build test clean install

dev:
	find content templates -name '*.md' -o -name '*.html' | entr -r go run ./cmd/blog serve

build:
	go build -o blog ./cmd/blog

install:
	go install ./cmd/blog

test:
	go test ./...

clean:
	rm -f blog
