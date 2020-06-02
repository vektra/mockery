FROM golang:1.14-alpine

COPY mockery /

# allow mockery to be accessible from $PATH
RUN ln -s /mockery /usr/local/bin

ENTRYPOINT ["/mockery"]
