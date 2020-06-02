FROM golang:1.14-alpine as builder

COPY ./ /mockery
RUN cd /mockery && go install ./...

FROM golang:1.14-alpine

COPY --from=builder /go/bin/mockery /

ENTRYPOINT ["/mockery"]
