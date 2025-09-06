# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ipssl-client .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata docker-cli

# Create non-root user and docker group
RUN addgroup -g 1001 -S appgroup && \
    addgroup -S docker && \
    adduser -u 1001 -S appuser -G appgroup && \
    adduser appuser docker

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/ipssl-client .

# Create required directories
RUN mkdir -p /ipssl /usr/share/caddy/.well-known/pki-validation && \
    chown -R appuser:appgroup /ipssl /usr/share/caddy

# Switch to non-root user
USER appuser

# Expose ports (if needed for health checks)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep ipssl-client || exit 1

# Run the application
CMD ["./ipssl-client"]
