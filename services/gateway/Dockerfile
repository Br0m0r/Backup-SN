ARG GO_VERSION=1.24.13
ARG ALPINE_VERSION=3.22
ARG ALPINE_RUNTIME_VERSION=3.22.2

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY services/common/ ./services/common/
COPY services/gateway/ ./services/gateway/

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o gateway ./services/gateway

FROM alpine:${ALPINE_RUNTIME_VERSION}

RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 10001 app && \
    adduser -S -D -H -u 10001 -G app app

WORKDIR /app
COPY --from=builder --chown=app:app /build/gateway ./gateway

USER 10001:10001
EXPOSE 8080
STOPSIGNAL SIGTERM

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["./gateway"]
