FROM alpine:latest

COPY bin/flarebot /usr/bin/flarebot

ENTRYPOINT ["/usr/bin/flarebot"]
