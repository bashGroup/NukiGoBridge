FROM golang:1.14

# Build and install
WORKDIR /go/src/nukibridge
COPY . .
RUN go get -d -v ./cmd/nukibridge
RUN go install -v ./cmd/nukibridge
RUN rm -rf /go/src /go/pkg

# Prepare
RUN mkdir -p /config
WORKDIR /config
VOLUME /config
ENV NUKI_CONFIGPATH /config

CMD ["nukibridge"]
