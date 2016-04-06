all: fmt test install

fmt:
	go fmt ./...

test:
	go test ./...

install:
	go install ./...
