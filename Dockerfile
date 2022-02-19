FROM golang:1.16-alpine

RUN apk add --no-cache git

RUN         go mod download
RUN         go build -o app
ENTRYPOINT  ["./app"]