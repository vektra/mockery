SHELL=bash

all: clean fmt test fixture install docker integration

clean:
	rm -rf mocks

fmt:
	go fmt ./...

test:
	go test ./...

fixture:
	mockery --print --dir pkg/fixtures --name RequesterVariadic > pkg/fixtures/mocks/requester_variadic.go
	mockery --print --dir pkg/fixtures --name RequesterVariadic --unroll-variadic=False > pkg/fixtures/mocks/requester_variadic_one_arg.go

install:
	go install ./...

docker:
	docker build -t vektra/mockery .

integration: docker install
	./hack/run-e2e.sh
