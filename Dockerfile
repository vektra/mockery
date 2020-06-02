FROM golang:1.14-alpine

# Let mockery write here even if container is run with --user $(id -u):$(id -g)
RUN mkdir /.cache && chmod 777 -R /.cache

COPY mockery /

ENTRYPOINT ["/mockery"]
