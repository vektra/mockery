SHELL=bash

all: clean fmt test fixture install integration

clean:
	rm -rf mocks

fmt:
	go fmt ./...

test:
	go test ./...

fixture:
	mockery -print -dir mockery/fixtures -name RequesterVariadic > mockery/fixtures/mocks/requester_variadic.go

install:
	go install ./...

integration:
	./hack/run-e2e.sh
