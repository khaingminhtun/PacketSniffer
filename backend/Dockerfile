# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24

# --- Compile (CGO + libpcap for gopacket) ---
FROM golang:${GO_VERSION}-bookworm AS build
RUN apt-get update \
  && apt-get install -y --no-install-recommends gcc libc6-dev libpcap-dev \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /sniffer ./cmd/sniffer

# --- Runtime ---
FROM debian:bookworm-slim AS runtime
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates libpcap0.8 \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /sniffer /usr/local/bin/sniffer

EXPOSE 8080
ENTRYPOINT ["sniffer"]
