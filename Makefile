.PHONY: build install clean test

build:
	go build -o ccl ./cmd/ccl

install:
	go install ./cmd/ccl/

clean:
	rm -f ccl

test:
	go test ./...
