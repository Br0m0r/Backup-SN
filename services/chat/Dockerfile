# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO (required for sqlite3 and websocket)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the common package first
COPY services/common ./services/common

# Copy source code
COPY services/chat ./services/chat

# Build with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o chat-service ./services/chat

# Run stage
FROM alpine:latest
RUN apk --no-cache add sqlite

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/chat-service .

# Expose port
EXPOSE 8085

# Run the service
CMD ["./chat-service"]
