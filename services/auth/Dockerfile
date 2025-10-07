# Dockerfile for Auth Service
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files from root project
COPY ../../go.mod ../../go.sum ./
RUN go mod download

# Copy auth service source code
COPY . .

# Build the auth service binary
RUN go build -o auth-service .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy the auth service binary
COPY --from=builder /app/auth-service .

# Auth service runs on port 8081
EXPOSE 8081

CMD ["./auth-service"]