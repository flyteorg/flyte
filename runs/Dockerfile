# Build stage
FROM golang:1.25.6-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make build-base

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire repository
COPY . ./

# Build the service
WORKDIR /workspace/runs
RUN CGO_ENABLED=1 GOOS=linux go build -a -o runs-service ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /workspace/runs/runs-service .

# Copy default config
COPY --from=builder /workspace/runs/config.yaml .

# Expose the service port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/healthz || exit 1

# Run the service
ENTRYPOINT ["./runs-service"]
CMD ["--config", "config.yaml"]
