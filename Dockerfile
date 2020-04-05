# Build stage
FROM golang:1.14 as build-golang
WORKDIR /go/src/nukibridge
COPY . .
RUN go generate ./...
RUN CGO_ENABLED=0 go build -v ./cmd/nukibridge

# Prepare
FROM alpine:3.11
COPY --from=build-golang /go/src/nukibridge/nukibridge /usr/local/bin/nukibridge
RUN mkdir -p /config
WORKDIR /config
VOLUME /config
ENV NUKI_CONFIGPATH /config

CMD ["nukibridge"]
