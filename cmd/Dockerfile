# Dockerfile for Polymarket CLOB Sequencer
FROM golang:1.23.6-alpine AS builder

# Install required packages
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/

# Build the application
WORKDIR /app/cmd
RUN go build -o sequencer .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh sequencer

# Set working directory
WORKDIR /home/sequencer

# Copy binary from builder stage
COPY --from=builder /app/cmd/sequencer .

# Change ownership to sequencer user
RUN chown sequencer:sequencer sequencer

# Switch to non-root user
USER sequencer

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./sequencer"]
