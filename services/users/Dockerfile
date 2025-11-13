# Dockerfile for User Service
FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files from root project
COPY go.mod go.sum ./
RUN go mod download

# Copy shared packages (notification helpers)
COPY services/common/ ./services/common/

# Copy user service source code
COPY services/users/ ./services/users/

# Build the user service binary with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o user-service ./services/users

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy the user service binary
COPY --from=builder /app/user-service .

# User service runs on port 8082
EXPOSE 8082

CMD ["./user-service"]

