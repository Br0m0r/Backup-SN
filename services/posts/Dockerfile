# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO (required for sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY services/posts ./services/posts

# Build with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o post-service ./services/posts

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/post-service .

# Expose port
EXPOSE 8083

# Run the service
CMD ["./post-service"]
