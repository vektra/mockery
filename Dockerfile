FROM golang:1.14-alpine

COPY mockery /usr/local/bin

ENTRYPOINT ["/usr/local/bin/mockery"]
