PREFIX ?= $(HOME)/.local

.PHONY: build install clean test

build:
	go build -o ccl ./cmd/ccl

install: build
	install -Dm755 ccl $(PREFIX)/bin/ccl

clean:
	rm -f ccl

test:
	go test ./...
