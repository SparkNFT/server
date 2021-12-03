FROM golang:1.17-bullseye AS builder

WORKDIR /app

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . .
RUN apt-get update && \
    apt-get install -y build-essential && \
    go build -o ./server ./cmd/server

# -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
FROM debian:bullseye
LABEL maintainer="Nyk Ma <nykma@mask.io>"

WORKDIR /app

RUN mkdir /app/config && \
    apt-get update && \
    apt-get install -y ca-certificates

COPY --from=builder /app/server .

CMD ["/app/server", "-config", "/app/config/config.json", "-debug"]
