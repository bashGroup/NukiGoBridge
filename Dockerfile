# Build vue app
FROM node:11-alpine as build-vue
WORKDIR /app
COPY webapp/ /app/
RUN yarn install
RUN yarn build

# Build stage
FROM golang:1.14 as build-golang
WORKDIR /go/src/nukibridge
COPY . .
COPY --from=build-vue /app/dist/ ./assets/
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
