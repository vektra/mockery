FROM scratch

COPY mockery /

ENTRYPOINT ["/mockery"]
