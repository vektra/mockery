FROM golang:1.14-alpine

COPY mockery /

RUN ln -s /mockery /go/bin

ENTRYPOINT ["/mockery"]
