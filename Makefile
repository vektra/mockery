SHELL=bash

.PHONY: all
all: fmt mocks test install docker integration

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test: mocks
	go test ./...

mocks: $(shell find . -type f -name '*.go' -not -name '*_test.go')
	go run . --dir pkg/fixtures --output mocks/pkg/fixtures
	go run . --all=false --print --dir pkg/fixtures --name RequesterVariadic --structname RequesterVariadicOneArgument --unroll-variadic=False > mocks/pkg/fixtures/RequesterVariadicOneArgument.go
	go run . --all=false --print --dir pkg/fixtures --name ExpecterTest --structname ExpecterTestWithExpecter --with-expecter=True > mocks/pkg/fixtures/ExpecterTestWithExpecter.go
	@touch mocks

.PHONY: install
install:
	go install .

.PHONY: docker
docker:
	docker build -t vektra/mockery .

.PHONY: integration
integration: docker install
	./hack/run-e2e.sh

.PHONY: clean
clean:
	rm -rf mocks
