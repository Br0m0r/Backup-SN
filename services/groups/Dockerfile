# Build stage
FROM golang:1.24-alpine AS builder

# Install gcc for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go.mod and go.sum from root
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the group service binary
RUN CGO_ENABLED=1 go build -o group-service ./services/groups

# Run stage
FROM alpine:latest

# Install SQLite runtime library
RUN apk add --no-cache sqlite-libs

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/group-service .

# Expose port 8084
EXPOSE 8084

# Run the service
CMD ["./group-service"]
