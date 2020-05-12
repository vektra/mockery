FROM golang:1.14-alpine

COPY mockery /

ENTRYPOINT ["/mockery"]
