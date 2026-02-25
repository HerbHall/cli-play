.PHONY: build test lint run clean

build:
	go build -o cli-play.exe ./cmd/cli-play

test:
	go test ./...

lint:
	golangci-lint run ./...

run:
	go run ./cmd/cli-play

clean:
	rm -f cli-play.exe
