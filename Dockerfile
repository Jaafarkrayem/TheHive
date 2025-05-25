# Multi-stage build for Hexagonal Chain Node
FROM golang:1.23-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the hexnode binary
RUN go build -o bin/hexnode ./cmd/hexnode

# Final stage - minimal runtime image
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh hexchain

# Set working directory
WORKDIR /home/hexchain

# Copy binary from builder
COPY --from=builder /app/bin/hexnode .

# Copy configuration files
COPY --from=builder /app/internal/config ./config

# Change ownership to hexchain user
RUN chown -R hexchain:hexchain /home/hexchain

# Switch to non-root user
USER hexchain

# Expose ports
EXPOSE 30303 8545 8546

# Default command
CMD ["./hexnode"] 