# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy notification service code
COPY services/notifications/ ./services/notifications/

# Build the service
RUN go build -o notification-service ./services/notifications

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add sqlite

WORKDIR /app

# Copy the binary from build stage
COPY --from=builder /app/notification-service .

# Expose port
EXPOSE 8086

# Run the service
CMD ["./notification-service"]
