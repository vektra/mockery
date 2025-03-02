FROM golang:1.24-alpine as builder

RUN apk --update add --no-cache gcc musl-dev git openssh

COPY mockery /usr/local/bin

# Explicitly set a writable cache path when running --user=$(id -u):$(id -g)
# see: https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE /tmp/.cache

ENTRYPOINT ["/usr/local/bin/mockery"]
