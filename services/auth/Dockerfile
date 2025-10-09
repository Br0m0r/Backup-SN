# Dockerfile for Auth Service
FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO(this was added because there was an error compiling with sqlite,
# particularly: sqlite3.Error: no such module: sqlite3)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files from root project
COPY go.mod go.sum ./
RUN go mod download

# Copy auth service source code
COPY services/auth/ ./services/auth/

# Build the auth service binary with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o auth-service ./services/auth

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy the auth service binary
COPY --from=builder /app/auth-service .

# Auth service runs on port 8081
EXPOSE 8081

CMD ["./auth-service"]