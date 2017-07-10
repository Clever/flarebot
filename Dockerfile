FROM golang:1.8-alpine

COPY bin/flarebot /usr/bin/flarebot

ENTRYPOINT ["/usr/bin/flarebot"]
