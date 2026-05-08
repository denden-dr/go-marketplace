# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /app/bin/go-marketplace cmd/api/main.go

# Stage 2: Final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/go-marketplace ./go-marketplace

# Copy migrations
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

# Expose port (overridden by PORT env var usually, but 8080 is default)
EXPOSE 8080

# Run the application
ENTRYPOINT ["./go-marketplace"]
