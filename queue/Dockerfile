# Build stage
FROM golang:1.24.6-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /workspace

# Use repo root go.mod/go.sum for better caching in monorepo
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire repository (rely on .dockerignore to trim)
COPY . ./

# Build the queue service binary from the queue module
WORKDIR /workspace/queue
RUN CGO_ENABLED=0 GOOS=linux go build -a -o queue-service ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and default config
COPY --from=builder /workspace/queue/queue-service .
COPY --from=builder /workspace/queue/config.yaml .

EXPOSE 8089

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8089/healthz || exit 1

ENTRYPOINT ["./queue-service"]
CMD ["--config", "config.yaml"]

