SHELL=bash

.PHONY: all
all: fmt mocks test install docker integration

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: mocks
mocks:
	go run . 

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
