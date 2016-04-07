SHELL=bash

all: clean fmt test install integration

clean:
	rm -rf mocks

fmt:
	go fmt ./...

test:
	go test ./...

install:
	go install ./...

integration:
	rm -rf mocks
	${GOPATH}/bin/mockery -all -recursive ./...
	if [ ! -d "mocks" ]; then \
		echo "No Mock Dir Created"; \
		exit 1; \
	fi
	if [ ! -f "mocks/AsyncProducer.go" ]; then \
		echo "AsyncProducer.go not created"; \
		echo 1; \
	fi