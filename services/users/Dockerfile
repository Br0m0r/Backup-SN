# Dockerfile for User Service
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files from root project
COPY ../../go.mod ../../go.sum ./
RUN go mod download

# Copy user service source code
COPY . .

# Build the user service binary
RUN go build -o user-service .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy the user service binary
COPY --from=builder /app/user-service .

# User service runs on port 8082
EXPOSE 8082

CMD ["./user-service"]