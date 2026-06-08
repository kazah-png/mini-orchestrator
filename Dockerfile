FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o orchestrator ./cmd/orchestrator

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y iptables iproute2 util-linux && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/orchestrator /usr/local/bin/
CMD ["orchestrator"]