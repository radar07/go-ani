BINARY := go-ani

.PHONY: build install test lint clean

build:
	go build -o $(BINARY) .

install:
	go install .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf $(BINARY)
